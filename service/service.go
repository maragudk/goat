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

type StartOptions struct {
	Continue bool
}

func (s *Service) Start(ctx context.Context, r io.Reader, w io.Writer, opts StartOptions) error {
	if err := s.DB.Connect(); err != nil {
		return errors.Wrap(err, "error connecting to database")
	}

	if err := s.DB.MigrateUp(ctx); err != nil {
		return errors.Wrap(err, "error migrating database")
	}

	var conversation model.Conversation
	var err error
	if opts.Continue {
		conversation, err = s.DB.GetLatestConversation(ctx)
		if err != nil {
			return errors.Wrap(err, "error getting latest conversation")
		}
		_, _ = fmt.Fprintln(w, "Continuing conversation", conversation.ID)
	} else {
		conversation, err = s.DB.NewConversation(ctx)
		if err != nil {
			return errors.Wrap(err, "error creating conversation")
		}
	}

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

		if !speakerNameMatcher.MatchString(text) {
			_, _ = fmt.Fprint(w, "> ")

			continue
		}

		matches := speakerNameMatcher.FindStringSubmatch(turn.Content)
		name := matches[1]
		llmSpeaker, err := s.DB.GetSpeakerByName(ctx, name)
		if err != nil {
			if errors.Is(err, model.ErrorSpeakerNotFound) {
				_, _ = fmt.Fprintf(w, "Error: No speaker called %v.\n\n> ", name)
				continue
			}
			return errors.Wrap(err, "error getting speaker by name")
		}

		client, ok := clients[llmSpeaker.ID]
		if !ok {
			m, err := s.DB.GetSpeakerModel(ctx, llmSpeaker.ID)
			if err != nil {
				return errors.Wrap(err, "error getting speaker model")
			}

			client = llm.NewOpenAIClient(llm.NewOpenAIClientOptions{
				BaseURL: m.URL(),
				Model:   llm.Model(m.Name),
				Token:   m.Token(),
			})
			clients[llmSpeaker.ID] = client
		}

		_, _ = fmt.Fprintln(w)

		cd, err := s.DB.GetConversationDocument(ctx, conversation.ID)
		if err != nil {
			return errors.Wrap(err, "error getting conversation document")
		}

		var messages []llm.Message
		for _, t := range cd.Turns {
			s := cd.Speakers[t.SpeakerID]

			prefix := ""
			role := llm.MessageRoleUser

			// If this is a turn from the current LLM, don't prefix the content with a name, and let the role be assistant
			if s.ID == llmSpeaker.ID {
				role = llm.MessageRoleAssistant
			} else {
				prefix = fmt.Sprintf("%v: ", s.Name)
			}

			messages = append(messages, llm.Message{
				Content: prefix + t.Content,
				Name:    s.Name,
				Role:    role,
			})
		}

		var b strings.Builder
		multiW := io.MultiWriter(w, &b)
		if err := client.Prompt(ctx, model.GlobalPrompt+llmSpeaker.System, messages, multiW); err != nil {
			return err
		}

		turn, err = s.DB.SaveTurn(ctx, model.Turn{
			ConversationID: conversation.ID,
			SpeakerID:      llmSpeaker.ID,
			Content:        b.String(),
		})
		if err != nil {
			return errors.Wrap(err, "error saving model turn")
		}

		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprint(w, "> ")
	}
	return nil
}
