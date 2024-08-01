package cmdutils

import (
	"io"
	"os/exec"

	"github.com/go-logr/logr"
	"k8s.io/utils/ptr"
)

type Options struct {
	executor executor
	logger   logr.Logger

	workdir string
	shell   *string
	args    []string

	stderr io.Writer
	stdout io.Writer
}

func newOptions(opts ...Option) *Options {
	o := &Options{
		executor: exec.CommandContext,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

type Option func(*Options)

func WithExecutor(executor executor) Option {
	return func(o *Options) {
		o.executor = executor
	}
}

func WithLogger(logger logr.Logger) Option {
	return func(o *Options) {
		o.logger = logger
	}
}

func WithWorkdir(workdir string) Option {
	return func(o *Options) {
		o.workdir = workdir
	}
}

func WithArgs(args ...string) Option {
	return func(o *Options) {
		o.args = args
	}
}

func WithStdout(w io.Writer) Option {
	return func(o *Options) {
		o.stdout = w
	}
}

func WithStderr(w io.Writer) Option {
	return func(o *Options) {
		o.stderr = w
	}
}

func WithShellMode(shell string) Option {
	return func(o *Options) {
		o.shell = ptr.To(shell)
	}
}
