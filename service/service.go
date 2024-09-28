package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"maragu.dev/errors"

	"maragu.dev/goat/llm"
	"maragu.dev/goat/model"
	goosql "maragu.dev/goo/sql"

	"maragu.dev/goat/sql"
)

type Service struct {
	DB *sql.Database
}

type NewOptions struct {
	Path string
}

func New(opts NewOptions) *Service {
	h := goosql.NewHelper(goosql.NewHelperOptions{
		Path: filepath.Join(opts.Path, "goat.db"),
	})
	db := sql.NewDatabase(sql.NewDatabaseOptions{
		SQLHelper: h,
	})

	return &Service{
		DB: db,
	}
}

// speakerNameMatcher matches speaker names. See https://regex101.com/r/QhwE8m/latest
var speakerNameMatcher = regexp.MustCompile(`\B@(?P<name>\w+)`)

func (s *Service) Start(ctx context.Context, r io.Reader, w io.Writer) error {
	if err := s.DB.Connect(); err != nil {
		return errors.Wrap(err, "error connecting to database")
	}

	if err := s.DB.MigrateUp(ctx); err != nil {
		return errors.Wrap(err, "error migrating database")
	}

	conversation, err := s.DB.NewConversation(ctx)
	if err != nil {
		return errors.Wrap(err, "error creating conversation")
	}

	var messages []llm.Message
	clients := map[model.ID]*llm.OpenAIClient{}

	_, _ = fmt.Fprint(w, "> ")

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()

		turn, err := s.DB.SaveTurn(ctx, model.Turn{
			ConversationID: conversation.ID,
			SpeakerID:      model.MySpeakerID,
			Content:        text,
		})
		if err != nil {
			return errors.Wrap(err, "error saving user turn")
		}

		messages = append(messages, llm.Message{
			Content: turn.Content,
			Name:    "Me",
			Role:    llm.MessageRoleUser,
		})

		if !speakerNameMatcher.MatchString(text) {
			fmt.Print("> ")

			continue
		}

		matches := speakerNameMatcher.FindStringSubmatch(turn.Content)
		name := matches[1]
		speaker, err := s.DB.GetSpeakerByName(ctx, name)
		if err != nil {
			if errors.Is(err, model.ErrorSpeakerNotFound) {
				_, _ = fmt.Fprintf(w, "Error: No speaker called %v.\n\n> ", name)
				continue
			}
			return errors.Wrap(err, "error getting speaker by name")
		}

		client, ok := clients[speaker.ID]
		if !ok {
			m, err := s.DB.GetSpeakerModel(ctx, speaker.ID)
			if err != nil {
				return errors.Wrap(err, "error getting speaker model")
			}

			client = llm.NewOpenAIClient(llm.NewOpenAIClientOptions{
				BaseURL: m.URL(),
				Model:   llm.Model(m.Name),
				Token:   m.Token(),
			})
			clients[speaker.ID] = client
		}

		_, _ = fmt.Fprintln(w)

		var b strings.Builder
		multiW := io.MultiWriter(w, &b)
		if err := client.Prompt(ctx, speaker.System, messages, multiW); err != nil {
			_, _ = fmt.Fprintln(multiW, "\n\nError: ", err)
			return err
		}

		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w)

		turn, err = s.DB.SaveTurn(ctx, model.Turn{
			ConversationID: conversation.ID,
			SpeakerID:      speaker.ID,
			Content:        b.String(),
		})
		if err != nil {
			return errors.Wrap(err, "error saving model turn")
		}

		messages = append(messages, llm.Message{
			Content: b.String(),
			Name:    speaker.Name,
			Role:    llm.MessageRoleAssistant,
		})

		fmt.Print("> ")
	}
	return nil
}
