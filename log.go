package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

var (
	globalLogLevel = slog.Level(-10)

	defaultLogHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: globalLogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "level" {
				return replaceLogLevelAttr(a)
			}
			return a
		},
	})
)

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
