package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jisun/tistory-indexer/internal/app"
)

func main() {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	a, err := app.Build(ctx)
	if err != nil {
		slog.Error("failed to initialize", "err", err)
		os.Exit(1)
	}

	if err := a.Service.Run(ctx); err != nil {
		slog.Error("indexer failed", "err", err)
		os.Exit(1)
	}
}
