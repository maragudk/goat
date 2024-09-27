package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"maragu.dev/errors"
	"maragu.dev/snorkel"

	"maragu.dev/goat/service"
)

func main() {
	log := snorkel.New(snorkel.Options{})

	if err := start(log); err != nil {
		log.Event("Error starting", 1, "error", err)
	}
}

func start(log *snorkel.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Make a .goat directory in the user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "error getting home directory")
	}
	goatDir := filepath.Join(home, ".goat")
	if err := os.MkdirAll(goatDir, 0700); err != nil {
		return errors.Wrap(err, "error creating .goat directory")
	}

	s := service.New(service.NewOptions{
		Path: goatDir,
	})
	if err := s.Start(ctx, os.Stdin, os.Stdout); err != nil {
		return err
	}
	return nil
}
