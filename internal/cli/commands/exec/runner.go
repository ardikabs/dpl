package exec

import (
	"github.com/ardikabs/dpl/internal/cli/global"
	"github.com/ardikabs/dpl/internal/logger"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
)

func runner(params *parameters, log logr.Logger) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		logger.SetLevel(global.GetLogLevel())

		if err := params.ParseArgs(args); err != nil {
			return err
		}

		if err := params.Validate(); err != nil {
			return err
		}

		instance, err := newExecInstance(log, params)
		if err != nil {
			return err
		}

		return instance.Exec(cmd.Context())
	}
}
