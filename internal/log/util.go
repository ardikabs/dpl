package log

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

func newSLogHandler() slog.Handler {
	opts := &slog.HandlerOptions{
		Level: &defaultLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "level" {
				return replaceLogLevelAttr(a)
			}
			return a
		},
	}

	return slog.NewTextHandler(os.Stdout, opts)
}

func replaceLogLevelAttr(lvl slog.Attr) slog.Attr {
	lvl.Key = "v"
	lvl.Value = slog.AnyValue(replaceDebugLevel(lvl.Value.String()))
	return lvl
}

func replaceDebugLevel(l string) string {
	after, found := strings.CutPrefix(l, "DEBUG")
	if !found {
		return l
	}

	level, err := strconv.Atoi(after)
	if err != nil {
		return fmt.Sprintf("DEBUG(4)")
	}

	return fmt.Sprintf("DEBUG(%d)", 4-level)
}
