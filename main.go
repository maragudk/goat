package main

import (
	"context"
	"embed"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"maragu.dev/errors"
	"maragu.dev/migrate"
	"maragu.dev/snorkel"

	"maragu.dev/goo/sql"
)

//go:embed sql/migrations
var migrations embed.FS

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

	h := sql.NewHelper(sql.NewHelperOptions{
		Path: filepath.Join(goatDir, "goat.db"),
	})
	if err := h.Connect(); err != nil {
		return errors.Wrap(err, "error connecting to database")
	}

	subFS, err := fs.Sub(migrations, "sql/migrations")
	if err != nil {
		return errors.Wrap(err, "error getting sub filesystem")
	}

	if err := migrate.Up(ctx, h.DB.DB, subFS); err != nil {
		return errors.Wrap(err, "error migrating")
	}

	return nil
}
