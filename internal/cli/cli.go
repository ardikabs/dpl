package cli

import (
	"os"

	"github.com/ardikabs/dpl/internal/cli/commands/exec"
	"github.com/ardikabs/dpl/internal/cli/commands/version"
	"github.com/ardikabs/dpl/internal/cli/global"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
)

func New(log logr.Logger) *cobra.Command {
	var cmd = &cobra.Command{
		Use:  "dpl",
		Long: "dpl is a CLI companion for managing the deployment process of the application to the Kubernetes cluster",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}

	global.AttachFlags(cmd.PersistentFlags())

	cmd.AddCommand(version.NewCommand())
	cmd.AddCommand(exec.NewCommand(log))
	return cmd
}
