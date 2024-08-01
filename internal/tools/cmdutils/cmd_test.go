package cmdutils_test

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/ardikabs/dpl/internal/tools/cmdutils"
	"github.com/stretchr/testify/require"
)

var fakeExecutor = func(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{
		"-test.run=TestShellProcessSuccess",
		"--",
		command,
	}

	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_TEST_PROCESS=1"}
	return cmd
}

func TestExec(t *testing.T) {
	t.Run("ideal task execution", func(t *testing.T) {
		err := cmdutils.Exec(context.TODO(), "ping", cmdutils.WithExecutor(fakeExecutor))
		require.NoError(t, err)
	})
}
