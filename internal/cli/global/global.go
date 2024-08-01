package global

import (
	"github.com/spf13/pflag"
)

var global = &flags{}

type flags struct {
	fs *pflag.FlagSet
}

func (f *flags) isUnset() bool {
	return f.fs == nil
}
