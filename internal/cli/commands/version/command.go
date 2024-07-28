package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   string
	GitCommit string
)

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:          "version",
		Short:        "Print the version number of dpl",
		Long:         `All software has versions. This is dpl version`,
		Example:      `$ dpl version`,
		SilenceUsage: false,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if len(Version) == 0 {
			fmt.Println("Version: dev")
		} else {
			fmt.Println("Version:", Version)
		}
		fmt.Println("Git Commit:", GitCommit)
	}

	return cmd
}
