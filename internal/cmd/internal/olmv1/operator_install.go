package olmv1

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/operator-framework/kubectl-operator/internal/cmd/internal/log"
	v1action "github.com/operator-framework/kubectl-operator/internal/pkg/v1/action"
	"github.com/operator-framework/kubectl-operator/pkg/action"
)

func NewOperatorInstallCmd(cfg *action.Configuration) *cobra.Command {
	i := v1action.NewOperatorInstall(cfg)
	i.Logf = log.Printf

	cmd := &cobra.Command{
		Use:   "operator <operator_name>",
		Short: "Install an operator",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			i.Package = args[0]
			_, err := i.Run(cmd.Context())
			if err != nil {
				log.Fatalf("failed to install operator: %v", err)
			}
			log.Printf("operator %q created", i.Package)
		},
	}
	bindOperatorInstallFlags(cmd.Flags(), i)

	return cmd
}

func bindOperatorInstallFlags(fs *pflag.FlagSet, i *v1action.OperatorInstall) {
	fs.StringVarP(&i.Namespace.Name, "namespace", "n", "", "namespace to install the operator in")
	fs.StringVarP(&i.Package, "package", "p", "", "package name of the operator to install")
	fs.StringSliceVarP(&i.Channels, "channels", "c", []string{}, "upgrade channels from which to resolve bundles")
	fs.StringVarP(&i.Version, "version", "v", "", "version (or version range) from which to resolve bundles")
	fs.StringVarP(&i.ServiceAccount, "service-account", "s", "default", "service account to use for the extension installation")
	fs.StringToStringVarP(&i.CatalogSelector.MatchLabels, "labels", "l", map[string]string{}, "labels that will be used to select catalog")
	fs.BoolVarP(&i.UnsafeCreateClusterRoleBinding, "unsafe-create-cluster-role-binding", "u", false, "create a cluster role binding for the operator's service account")
	fs.DurationVarP(&i.CleanupTimeout, "cleanup-timeout", "d", time.Minute, "the amount of time to wait before cancelling cleanup after a failed creation attempt")
}
