package main

import (
	"log/slog"
	"os"
)

func main() {
	options := &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 && attr.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return attr
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, options))
	logger.Info("借书完成", "book", "Go", "days", 7)
}
