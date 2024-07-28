package global

import "github.com/spf13/pflag"

const (
	LOG_VERBOSITY = "v"
)

func SetFlags(flagset *pflag.FlagSet) {
	flagset.IntP(LOG_VERBOSITY, "v", 0, "Number for the log level verbosity")
}

func GetLogLevel(flagset *pflag.FlagSet) int {
	v, err := flagset.GetInt(LOG_VERBOSITY)
	if err != nil {
		return 0
	}

	return v
}
