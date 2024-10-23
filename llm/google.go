package llm

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"maragu.dev/snorkel"
)

const (
	ModelGemini_1_5_Flash = Model("gemini-1.5-flash")
)

type GoogleClient struct {
	client *genai.Client
	model  Model
}

type NewGoogleClientOptions struct {
	Log   *snorkel.Logger
	Model Model
	Token string
}

func NewGoogleClient(opts NewGoogleClientOptions) *GoogleClient {
	if opts.Log == nil {
		opts.Log = snorkel.NewDiscard()
	}

	client, err := genai.NewClient(context.Background(), option.WithAPIKey(opts.Token))
	if err != nil {
		panic(err)
	}

	return &GoogleClient{
		client: client,
		model:  opts.Model,
	}
}

func (c *GoogleClient) Prompt(ctx context.Context, system string, messages []Message, w io.Writer) error {
	model := c.client.GenerativeModel(c.model.String())

	if system != "" {
		model.SystemInstruction = genai.NewUserContent(genai.Text(system))
	}

	cs := model.StartChat()

	var history []*genai.Content
	for i, m := range messages {
		var role string
		switch m.Role {
		case MessageRoleUser:
			role = "user"
		case MessageRoleAssistant:
			role = "model"
		}
		history = append(history, &genai.Content{
			Parts: []genai.Part{genai.Text(m.Content)},
			Role:  role,
		})
		if i == len(messages)-2 {
			break
		}
	}
	cs.History = history

	m := messages[len(messages)-1]
	content := genai.Text(m.Content)
	iter := cs.SendMessageStream(ctx, content)
	for {
		res, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return err
		}
		for _, cand := range res.Candidates {
			if cand.Content == nil {
				continue
			}
			for _, part := range cand.Content.Parts {
				if _, err := fmt.Fprint(w, part); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
