package sql_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"maragu.dev/is"

	"maragu.dev/goat/model"
	"maragu.dev/goat/sql"
	goosql "maragu.dev/goo/sql"
)

func TestDatabase_SaveTurn(t *testing.T) {
	t.Run("can save a turn", func(t *testing.T) {
		db := newDB(t)

		c, err := db.NewConversation(context.Background())
		is.NotError(t, err)

		turn, err := db.SaveTurn(context.Background(), model.Turn{
			ConversationID: c.ID,
			SpeakerID:      model.MySpeakerID,
			Content:        "Testing, testing.",
		})
		is.NotError(t, err)
		is.Equal(t, c.ID, turn.ConversationID)
		is.Equal(t, model.MySpeakerID, turn.SpeakerID)
		is.Equal(t, "Testing, testing.", turn.Content)
	})

	t.Run("updates the last turn if the speaker is the same", func(t *testing.T) {
		db := newDB(t)

		c, err := db.NewConversation(context.Background())
		is.NotError(t, err)

		turn1, err := db.SaveTurn(context.Background(), model.Turn{
			ConversationID: c.ID,
			SpeakerID:      model.MySpeakerID,
			Content:        "Testing, testing.",
		})
		is.NotError(t, err)

		turn2, err := db.SaveTurn(context.Background(), model.Turn{
			ConversationID: c.ID,
			SpeakerID:      model.MySpeakerID,
			Content:        "Really.",
		})
		is.NotError(t, err)

		is.Equal(t, turn1.ID, turn2.ID)
		is.Equal(t, "Testing, testing.\nReally.", turn2.Content)
	})

	t.Run("errors if no such speaker", func(t *testing.T) {
		t.Skip()
		db := newDB(t)

		c, err := db.NewConversation(context.Background())
		is.NotError(t, err)

		_, err = db.SaveTurn(context.Background(), model.Turn{
			ConversationID: c.ID,
			SpeakerID:      "s_doesnotexist",
			Content:        "Testing, testing.",
		})
		is.Error(t, model.ErrorSpeakerNotFound, err)
	})
}

func TestDatabase_GetSpeakerByName(t *testing.T) {
	t.Run("gets the speaker by name", func(t *testing.T) {
		db := newDB(t)

		s, err := db.GetSpeakerByName(context.Background(), "me")
		is.NotError(t, err)
		is.Equal(t, model.MySpeakerID, s.ID)
	})

	t.Run("returns an error if no such speaker", func(t *testing.T) {
		db := newDB(t)

		_, err := db.GetSpeakerByName(context.Background(), "doesnotexist")
		is.Error(t, model.ErrorSpeakerNotFound, err)
	})
}

func TestDatabase_GetSpeakerModel(t *testing.T) {
	t.Run("gets the speaker model", func(t *testing.T) {
		db := newDB(t)

		m, err := db.GetSpeakerModel(context.Background(), model.MySpeakerID)
		is.NotError(t, err)
		is.Equal(t, model.ModelTypeBrain, m.Type)
	})

	t.Run("errors if model not found", func(t *testing.T) {
		db := newDB(t)

		_, err := db.GetSpeakerModel(context.Background(), "s_doesnotexist")
		is.Error(t, model.ErrorModelNotFound, err)
	})
}

func TestDatabase_GetSpeakers(t *testing.T) {
	t.Run("gets speakers and their model names", func(t *testing.T) {
		db := newDB(t)

		ss, err := db.GetSpeakers(context.Background())
		is.NotError(t, err)
		is.Equal(t, 8, len(ss))
		is.Equal(t, "claude", ss[0].Name)
		is.Equal(t, "claude-3-5-sonnet-20240620", ss[0].ModelName)
	})
}

func newDB(t *testing.T) *sql.Database {
	t.Helper()

	cleanup(t)
	t.Cleanup(func() {
		cleanup(t)
	})

	h := goosql.NewHelper(goosql.NewHelperOptions{
		Path: "test.db",
	})
	db := sql.NewDatabase(sql.NewDatabaseOptions{
		SQLHelper: h,
	})
	if err := db.Connect(); err != nil {
		t.Fatal(err)
	}
	if err := db.MigrateUp(context.Background()); err != nil {
		t.Fatal(err)
	}

	return db
}

func cleanup(t *testing.T) {
	t.Helper()

	files, err := filepath.Glob("test.db*")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			t.Fatal(err)
		}
	}
}
