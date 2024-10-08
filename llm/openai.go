package llm

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"maragu.dev/errors"
	"maragu.dev/snorkel"
)

const (
	ModelLlama3_2_1B = Model("llama3.2-1b")
	ModelLlama3_2_3B = Model("llama3.2-3b")
	ModelGPT4o       = Model(openai.GPT4o)
)

type OpenAIClient struct {
	client *openai.Client
	log    *snorkel.Logger
	model  Model
}

type NewOpenAIClientOptions struct {
	BaseURL string
	Log     *snorkel.Logger
	Model   Model
	Token   string
}

func NewOpenAIClient(opts NewOpenAIClientOptions) *OpenAIClient {
	if opts.Log == nil {
		opts.Log = snorkel.NewDiscard()
	}

	config := openai.DefaultConfig(opts.Token)

	if opts.BaseURL != "" {
		config.BaseURL = opts.BaseURL
	}

	return &OpenAIClient{
		client: openai.NewClientWithConfig(config),
		log:    opts.Log,
		model:  opts.Model,
	}
}

func (c *OpenAIClient) Prompt(ctx context.Context, system string, messages []Message, w io.Writer) error {
	var ccms []openai.ChatCompletionMessage
	if system != "" {
		ccms = append(ccms, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: system,
		})
	}

	for _, m := range messages {
		ccms = append(ccms, openai.ChatCompletionMessage{
			Content: m.Content,
			Name:    m.Name,
			Role:    string(m.Role),
		})
	}

	stream, err := c.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    c.model.String(),
		Messages: ccms,
		Stream:   true,
	})
	if err != nil {
		return errors.Wrap(err, "error creating openai chat completion stream")
	}
	defer func() {
		_ = stream.Close()
	}()

	for {
		res, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			e := &openai.APIError{}
			if errors.As(err, &e) {
				switch e.HTTPStatusCode {
				case http.StatusTooManyRequests, http.StatusInternalServerError:
					time.Sleep(time.Second)
					continue
				default:
					return errors.Wrap(err, "openai api error receiving chat completion stream")
				}
			}
			return errors.Wrap(err, "error receiving openai chat completion stream")
		}

		content := res.Choices[0].Delta.Content
		content = strings.ReplaceAll(content, "<|eot_id|>", "")

		if _, err := w.Write([]byte(content)); err != nil {
			return err
		}
	}
}
