package llm

import (
	"context"
	"io"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"maragu.dev/snorkel"
)

const (
	ModelClaude_3_Haiku    = Model(anthropic.ModelClaude_3_Haiku_20240307)
	ModelClaude_3_5_Sonnet = Model(anthropic.ModelClaude_3_5_Sonnet_20240620)
)

type AnthropicClient struct {
	client *anthropic.Client
	log    *snorkel.Logger
	model  Model
}

type NewAnthropicClientOptions struct {
	Log   *snorkel.Logger
	Model Model
	Token string
}

func NewAnthropicClient(opts NewAnthropicClientOptions) *AnthropicClient {
	if opts.Log == nil {
		opts.Log = snorkel.NewDiscard()
	}

	return &AnthropicClient{
		client: anthropic.NewClient(option.WithAPIKey(opts.Token)),
		log:    opts.Log,
		model:  opts.Model,
	}
}

func (c *AnthropicClient) Prompt(ctx context.Context, system string, messages []Message, w io.Writer) error {
	var mps []anthropic.MessageParam

	for _, m := range messages {
		switch m.Role {
		case MessageRoleUser:
			mps = append(mps, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Name+": "+m.Content)))
		case MessageRoleAssistant:
			mps = append(mps, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		}
	}

	params := anthropic.MessageNewParams{
		MaxTokens: anthropic.Int(4096),
		Messages:  anthropic.F(mps),
		Model:     anthropic.F(anthropic.ModelClaude_3_5_Sonnet_20240620),
	}
	if system != "" {
		params.System = anthropic.F([]anthropic.TextBlockParam{anthropic.NewTextBlock(system)})
	}
	stream := c.client.Messages.NewStreaming(ctx, params)
	defer func() {
		_ = stream.Close()
	}()

	var message anthropic.Message
	for stream.Next() {
		event := stream.Current()
		if err := message.Accumulate(event); err != nil {
			return err
		}

		switch delta := event.Delta.(type) {
		case anthropic.ContentBlockDeltaEventDelta:
			if delta.Text != "" {
				if _, err := w.Write([]byte(delta.Text)); err != nil {
					return err
				}
			}
		}
	}
	if err := stream.Err(); err != nil {
		return err
	}

	return nil
}
