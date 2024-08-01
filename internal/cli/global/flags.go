package global

import "github.com/spf13/pflag"

const (
	LOG_VERBOSITY = "v"
)

func AttachFlags(flagset *pflag.FlagSet) {
	flagset.IntP(LOG_VERBOSITY, "v", 0, "Number for the log level verbosity")

	global.fs = flagset
}

func GetLogLevel() int {
	if global.isUnset() {
		return 0
	}

	v, err := global.fs.GetInt(LOG_VERBOSITY)
	if err != nil {
		return 0
	}

	return v
}
