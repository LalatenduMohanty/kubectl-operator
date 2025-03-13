package cmd

import (
	"github.com/spf13/cobra"

	"github.com/operator-framework/kubectl-operator/internal/cmd/internal/olmv1"
	"github.com/operator-framework/kubectl-operator/pkg/action"
)

func newOlmV1Cmd(cfg *action.Configuration) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "olmv1",
		Short: "Manage operators via OLMv1 in a cluster from the command line",
		Long:  "Manage operators via OLMv1 in a cluster from the command line.",
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Display one or many resource(s)",
		Long:  "Display one or many resource(s)",
	}
	getCmd.AddCommand(
		olmv1.NewOperatorInstalledGetCmd(cfg),
		olmv1.NewCatalogInstalledGetCmd(cfg),
	)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a resource",
		Long:  "Create a resource",
	}
	createCmd.AddCommand(olmv1.NewCatalogCreateCmd(cfg))

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a resource",
		Long:  "Delete a resource",
	}
	deleteCmd.AddCommand(
		olmv1.NewCatalogDeleteCmd(cfg),
		olmv1.NewOperatorUninstallCmd(cfg),
	)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update a resource",
		Long:  "Update a resource",
	}
	updateCmd.AddCommand(
		olmv1.NewOperatorUpdateCmd(cfg),
	)

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install a resource",
		Long:  "Install a resource",
	}
	installCmd.AddCommand(olmv1.NewOperatorInstallCmd(cfg))

	cmd.AddCommand(
		installCmd,
		getCmd,
		createCmd,
		deleteCmd,
		updateCmd,
	)

	return cmd
}
