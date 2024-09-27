package sql_test

import (
	"context"
	"testing"

	"maragu.dev/is"

	"maragu.dev/goo/sqltest"
)

func TestDatabase_Migrate(t *testing.T) {
	t.Run("can migrate down and back up", func(t *testing.T) {
		h := sqltest.NewHelper(t)

		err := h.MigrateDown(context.Background())
		is.NotError(t, err)

		err = h.MigrateUp(context.Background())
		is.NotError(t, err)
	})
}
