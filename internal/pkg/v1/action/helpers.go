package action

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	olmv1 "github.com/operator-framework/operator-controller/api/v1"
)

const pollInterval = 250 * time.Millisecond

func objectKeyForObject(obj client.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}

func waitUntilCatalogStatusCondition(
	ctx context.Context,
	cl getter,
	catalog *olmv1.ClusterCatalog,
	conditionType string,
	conditionStatus metav1.ConditionStatus,
) error {
	opKey := objectKeyForObject(catalog)
	return wait.PollUntilContextCancel(ctx, pollInterval, true, func(conditionCtx context.Context) (bool, error) {
		if err := cl.Get(conditionCtx, opKey, catalog); err != nil {
			return false, err
		}

		if slices.ContainsFunc(catalog.Status.Conditions, func(cond metav1.Condition) bool {
			return cond.Type == conditionType && cond.Status == conditionStatus
		}) {
			return true, nil
		}
		return false, nil
	})
}

func waitUntilOperatorStatusCondition(
	ctx context.Context,
	cl getter,
	operator *olmv1.ClusterExtension,
	conditionType string,
	conditionStatus metav1.ConditionStatus,
) error {
	opKey := objectKeyForObject(operator)
	return wait.PollUntilContextCancel(ctx, pollInterval, true, func(conditionCtx context.Context) (bool, error) {
		if err := cl.Get(conditionCtx, opKey, operator); err != nil {
			return false, err
		}

		if slices.ContainsFunc(operator.Status.Conditions, func(cond metav1.Condition) bool {
			return cond.Type == conditionType && cond.Status == conditionStatus
		}) {
			return true, nil
		}
		return false, nil
	})
}

func deleteWithTimeout(cl deleter, obj client.Object, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := cl.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	return nil
}

func waitForDeletion(ctx context.Context, cl getter, objs ...client.Object) error {
	for _, obj := range objs {
		lowerKind := strings.ToLower(obj.GetObjectKind().GroupVersionKind().Kind)
		key := objectKeyForObject(obj)
		if err := wait.PollUntilContextCancel(ctx, pollInterval, true, func(conditionCtx context.Context) (bool, error) {
			if err := cl.Get(conditionCtx, key, obj); apierrors.IsNotFound(err) {
				return true, nil
			} else if err != nil {
				return false, err
			}
			return false, nil
		}); err != nil {
			return fmt.Errorf("wait for %s %q deleted: %v", lowerKind, key.Name, err)
		}
	}
	return nil
}

func patchObject(ctx context.Context, cl client.Client, obj interface{}) error {
	var (
		u   unstructured.Unstructured
		err error
	)
	u.Object, err = runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}
	return cl.Patch(ctx, &u, client.Apply)
}

func deleteAndWait(ctx context.Context, cl client.Client, objs ...client.Object) error {
	var (
		wg   sync.WaitGroup
		errs = make([]error, len(objs))
	)
	for i := range objs {
		wg.Add(1)
		go func(objectIndex int) {
			defer wg.Done()
			obj := objs[objectIndex]
			gvk := obj.GetObjectKind().GroupVersionKind()
			if gvk.Empty() {
				gvks, unversioned, err := cl.Scheme().ObjectKinds(obj)
				if err == nil && !unversioned && len(gvks) > 0 {
					gvk = gvks[0]
				}
			}
			lowerKind := strings.ToLower(gvk.Kind)
			key := client.ObjectKeyFromObject(obj)

			err := cl.Delete(ctx, obj)
			if err != nil && !apierrors.IsNotFound(err) {
				errs[i] = fmt.Errorf("delete %s %q: %v", lowerKind, key.Name, err)
				return
			}

			if err := wait.PollUntilContextCancel(ctx, 250*time.Millisecond, true, func(conditionCtx context.Context) (bool, error) {
				if err := cl.Get(conditionCtx, key, obj); apierrors.IsNotFound(err) {
					return true, nil
				} else if err != nil {
					return false, err
				}
				return false, nil
			}); err != nil {
				errs[i] = fmt.Errorf("wait for %s %q deleted: %v", lowerKind, key.Name, err)
				return
			}
		}(i)
	}
	wg.Wait()
	return errors.Join(errs...)
}
