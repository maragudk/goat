package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/muesli/termenv"
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
	Prompt   string
}

type prompter interface {
	Prompt(ctx context.Context, system string, messages []llm.Message, w io.Writer) error
}

func (s *Service) Start(ctx context.Context, r io.Reader, w io.Writer, opts StartOptions) error {
	interactive := true
	if opts.Prompt != "" {
		if !speakerNameMatcher.MatchString(opts.Prompt) {
			return errors.New("no speaker mentioned in prompt")
		}

		interactive = false
		opts.Continue = false
		r = strings.NewReader(opts.Prompt)
	}

	var conversation model.Conversation
	clients := map[model.ID]prompter{}

	mySpeaker, err := s.DB.GetSpeaker(ctx, model.MySpeakerID)
	if err != nil {
		return errors.Wrap(err, "error getting my speaker")
	}

	output := termenv.NewOutput(w)
	w = output

	// If we're continuing a conversation, print the conversation so far
	if opts.Continue {
		conversation, err = s.DB.GetLatestConversation(ctx)
		if err != nil {
			return errors.Wrap(err, "error getting latest conversation")
		}

		if conversation.Topic != "" {
			output.SetWindowTitle(conversation.Topic)
		}

		cd, err := s.DB.GetConversationDocument(ctx, conversation.ID)
		if err != nil {
			return errors.Wrap(err, "error getting conversation document")
		}
		for _, t := range cd.Turns {
			s := cd.Speakers[t.SpeakerID]
			printAvatar(w, s)
			_, _ = fmt.Fprintln(w, t.Content)
			printTurnSeparator(w)
		}
	}

	if interactive {
		printAvatar(w, mySpeaker)
	}

	var summarizer prompter

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())

		// Only initalize the conversation once we have some text
		var err error
		if conversation.ID == "" {
			conversation, err = s.DB.NewConversation(ctx)
			if err != nil {
				return errors.Wrap(err, "error creating conversation")
			}
		}

		turn, err := s.DB.SaveTurn(ctx, model.Turn{
			ConversationID: conversation.ID,
			SpeakerID:      model.MySpeakerID,
			Content:        text,
		})
		if err != nil {
			return errors.Wrap(err, "error saving user turn")
		}

		if !speakerNameMatcher.MatchString(text) {
			continue
		}

		matches := speakerNameMatcher.FindStringSubmatch(turn.Content)
		name := matches[1]
		llmSpeaker, err := s.DB.GetSpeakerByName(ctx, name)
		if err != nil {
			if errors.Is(err, model.ErrorSpeakerNotFound) {
				_, _ = fmt.Fprintf(w, "Error: No speaker called %v.\n\n", name)
				printAvatar(w, mySpeaker)
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

			client = newClientFromModel(m)

			clients[llmSpeaker.ID] = client

			if summarizer == nil {
				summarizer = client
			}
		}

		if interactive {
			printTurnSeparator(w)
		}

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

		printAvatar(w, llmSpeaker)

		var b strings.Builder
		multiW := io.MultiWriter(w, &b)

		prompt := model.CreateGlobalPrompt(llmSpeaker.Name) + llmSpeaker.System
		if err := client.Prompt(ctx, prompt, messages, multiW); err != nil {
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

		if summarizer != nil {
			messages = append(messages, llm.Message{
				Content: b.String(),
				Name:    llmSpeaker.Name,
				Role:    llm.MessageRoleAssistant,
			})

			messages = append(messages, llm.Message{
				Content: "Summarize the above conversation.",
				Role:    llm.MessageRoleUser,
			})

			var summary strings.Builder
			if err := summarizer.Prompt(ctx, model.CreateSummarizerPrompt(llmSpeaker.Name), messages, &summary); err != nil {
				return errors.Wrap(err, "error summarizing conversation")
			}

			if err := s.DB.SaveTopic(ctx, conversation.ID, summary.String()); err != nil {
				return errors.Wrap(err, "error saving conversation topic")
			}

			output.SetWindowTitle(summary.String())
		}

		if interactive {
			printTurnSeparator(w)
		}

		if !interactive {
			break
		}

		printAvatar(w, mySpeaker)
	}
	return nil
}

func newClientFromModel(m model.Model) prompter {
	var client prompter
	switch m.Type {
	case model.ModelTypeLlamaCPP, model.ModelTypeOpenAI, model.ModelTypeGroq, model.ModelTypeHuggingFace:
		client = llm.NewOpenAIClient(llm.NewOpenAIClientOptions{
			BaseURL: m.URL(),
			Model:   llm.Model(m.Name),
			Token:   m.Token(),
		})
	case model.ModelTypeAnthropic:
		client = llm.NewAnthropicClient(llm.NewAnthropicClientOptions{
			Model: llm.Model(m.Name),
			Token: m.Token(),
		})
	}
	return client
}

func printAvatar(w io.Writer, s model.Speaker) {
	_, _ = fmt.Fprint(w, s.Avatar()+": ")
}

func printTurnSeparator(w io.Writer) {
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w)
}

func (s *Service) ConnectAndMigrate(ctx context.Context) error {
	if err := s.DB.Connect(); err != nil {
		return errors.Wrap(err, "error connecting to database")
	}

	if err := s.DB.MigrateUp(ctx); err != nil {
		return errors.Wrap(err, "error migrating database")
	}
	return nil
}
