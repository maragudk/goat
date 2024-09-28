package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"maragu.dev/env"
	"maragu.dev/errors"

	"maragu.dev/goat/service"
)

func main() {
	if err := start(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
	}
}

func start() error {
	_ = env.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	dir := env.GetStringOrDefault("GOAT_DIR", "")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return errors.Wrap(err, "error getting home directory")
		}
		dir = filepath.Join(home, ".goat")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return errors.Wrap(err, "error creating .goat directory")
	}

	s := service.New(service.NewOptions{
		Path: dir,
	})
	if err := s.Start(ctx, os.Stdin, os.Stdout); err != nil {
		return err
	}
	return nil
}
