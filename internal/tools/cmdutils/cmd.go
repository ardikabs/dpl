package cmdutils

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/ardikabs/dpl/internal/errs"
	"k8s.io/utils/ptr"
)

var (
	ErrCommandFailed = fmt.Errorf("failed to execute command")
)

type executor func(ctx context.Context, cmd string, args ...string) *exec.Cmd

func Exec(ctx context.Context, cmd string, opts ...Option) error {
	o := newOptions(opts...)

	log := o.logger.WithName("cmd.Exec")

	command := o.executor(ctx, cmd, o.args...)
	command.Dir = o.workdir

	if o.shell != nil {
		shellExec := ptr.Deref(o.shell, "/bin/sh")
		command.Path = shellExec
		command.Args = append([]string{"-c", cmd}, command.Args...)

		log = log.WithValues("shell", shellExec)
	}

	log = log.WithValues(
		"cmd", command.String(),
		"workdir", o.workdir,
	)

	log.V(1).Info("executing command")

	var err error
	if o.stdout == nil || o.stderr == nil {
		var out []byte
		out, err = command.CombinedOutput()

		log = log.WithValues("output", string(out))
	} else {
		command.Stdout = o.stdout
		command.Stderr = o.stderr

		err = command.Run()
	}

	if err != nil {
		log.V(1).Info("command failed for some reason")
		return errs.Wrap(err, ErrCommandFailed)
	}

	log.Info("command executed successfully")
	return nil
}
