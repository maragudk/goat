package sql

import (
	"context"
	"embed"
	"io/fs"

	_ "github.com/mattn/go-sqlite3"
	"maragu.dev/errors"
	"maragu.dev/migrate"
	"maragu.dev/snorkel"

	"maragu.dev/goo/sql"
)

type Database struct {
	h   *sql.Helper
	log *snorkel.Logger
}

type NewDatabaseOptions struct {
	Log       *snorkel.Logger
	SQLHelper *sql.Helper
}

// NewDatabase with the given options.
// If no logger is provided, logs are discarded.
func NewDatabase(opts NewDatabaseOptions) *Database {
	if opts.Log == nil {
		opts.Log = snorkel.NewDiscard()
	}

	return &Database{
		log: opts.Log,
		h:   opts.SQLHelper,
	}
}

func (d *Database) Connect() error {
	if err := d.h.Connect(); err != nil {
		return err
	}
	return nil
}

//go:embed migrations
var migrations embed.FS

func (d *Database) MigrateUp(ctx context.Context) error {
	subFS, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return errors.Wrap(err, "error getting sub filesystem")
	}

	if err := migrate.Up(ctx, d.h.DB.DB, subFS); err != nil {
		return errors.Wrap(err, "error migrating")
	}

	return nil
}
