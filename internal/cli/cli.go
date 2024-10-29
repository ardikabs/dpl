package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ardikabs/dpl/internal/cli/commands/exec"
	"github.com/ardikabs/dpl/internal/cli/commands/version"
	"github.com/ardikabs/dpl/internal/cli/global"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	name := filepath.Base(os.Args[0])

	var cmd = &cobra.Command{
		Use:  name,
		Long: fmt.Sprintf("%s is a CLI companion for managing the deployment process of the application to the Kubernetes cluster", name),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}

	global.AttachFlags(cmd.PersistentFlags())

	cmd.AddCommand(version.NewCommand())
	cmd.AddCommand(exec.NewCommand())
	return cmd
}
