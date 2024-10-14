package sql

import (
	"context"
	"database/sql"

	"github.com/mattn/go-sqlite3"
	"maragu.dev/errors"

	"maragu.dev/goat/model"
	goosql "maragu.dev/goo/sql"
)

func (d *Database) NewConversation(ctx context.Context) (model.Conversation, error) {
	var c model.Conversation
	err := d.h.Get(ctx, &c, "insert into conversations default values returning *")
	return c, err
}

func (d *Database) GetLatestConversation(ctx context.Context) (model.Conversation, error) {
	var c model.Conversation
	err := d.h.Get(ctx, &c, "select * from conversations order by created desc limit 1")
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return c, model.ErrorConversationNotFound
	}
	return c, err
}

func (d *Database) GetConversationDocument(ctx context.Context, id model.ID) (model.ConversationDocument, error) {
	var cd model.ConversationDocument
	cd.Speakers = map[model.ID]model.Speaker{}

	err := d.h.InTransaction(ctx, func(tx *goosql.Tx) error {
		err := tx.Get(ctx, &cd.Conversation, `select * from conversations where id = ?`, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return model.ErrorConversationNotFound
			}
			return err
		}
		err = tx.Select(ctx, &cd.Turns, `select * from turns where conversationID = ? order by created`, id)
		if err != nil {
			return err
		}
		for _, t := range cd.Turns {
			s, ok := cd.Speakers[t.SpeakerID]
			if ok {
				continue
			}
			err = tx.Get(ctx, &s, `select * from speakers where id = ?`, t.SpeakerID)
			if err != nil {
				return err
			}
			cd.Speakers[s.ID] = s
		}
		return nil
	})
	return cd, err
}

func (d *Database) SaveTurn(ctx context.Context, t model.Turn) (model.Turn, error) {
	err := d.h.InTransaction(ctx, func(tx *goosql.Tx) error {
		var lastSpeakerID model.ID
		query := `select speakerID from turns where conversationID = ? order by created desc limit 1`
		if err := tx.Get(ctx, &lastSpeakerID, query, t.ConversationID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		if lastSpeakerID != t.SpeakerID {
			query = "insert into turns (conversationID, speakerID, content) values (?, ?, ?) returning *"
			if err := tx.Get(ctx, &t, query, t.ConversationID, t.SpeakerID, t.Content); err != nil {
				switch {
				case isForeignKeyConstraintError(err, "conversationID"):
					return model.ErrorConversationNotFound
				case isForeignKeyConstraintError(err, "speakerID"):
					return model.ErrorSpeakerNotFound
				default:
					return err
				}
			}
			return nil
		}

		query = `update turns set content = content || ? where id = (select id from turns where conversationID = ? order by created desc limit 1) returning *`
		if err := tx.Get(ctx, &t, query, "\n"+t.Content, t.ConversationID); err != nil {
			return err
		}
		return nil

	})
	return t, err
}

func (d *Database) GetSpeaker(ctx context.Context, id model.ID) (model.Speaker, error) {
	var s model.Speaker
	err := d.h.Get(ctx, &s, "select * from speakers where id = ?", id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return s, model.ErrorSpeakerNotFound
	}
	return s, err
}

func (d *Database) GetSpeakerByName(ctx context.Context, name string) (model.Speaker, error) {
	var s model.Speaker
	err := d.h.Get(ctx, &s, "select * from speakers where name = ?", name)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return s, model.ErrorSpeakerNotFound
	}
	return s, err
}

func (d *Database) GetSpeakers(ctx context.Context) ([]model.SpeakerModelPair, error) {
	var ss []model.SpeakerModelPair
	query := `
		select speakers.*, models.name as ModelName from speakers
			join models on models.id = speakers.modelID
		order by speakers.name`
	err := d.h.Select(ctx, &ss, query)
	return ss, err
}

func (d *Database) GetSpeakerModel(ctx context.Context, speakerID model.ID) (model.Model, error) {
	var m model.Model
	err := d.h.Get(ctx, &m, "select * from models where id = (select modelID from speakers where id = ?)", speakerID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return m, model.ErrorModelNotFound
	}
	return m, err
}

func (d *Database) GetModels(ctx context.Context) ([]model.Model, error) {
	var ms []model.Model
	err := d.h.Select(ctx, &ms, "select * from models order by name")
	return ms, err
}

func isForeignKeyConstraintError(err error, column string) bool {
	// TODO figure out if we can check the column
	var sqliteErr sqlite3.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}

	return errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintForeignKey)
}
