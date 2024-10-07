package llm_test

import (
	"context"
	"strings"
	"testing"

	"maragu.dev/env"
	"maragu.dev/is"

	"maragu.dev/goat/llm"
)

func TestAnthropicClient_Prompt(t *testing.T) {
	_ = env.Load("../.env.test.local")

	token := env.GetStringOrDefault("ANTHROPIC_TOKEN", "123")

	tests := []struct {
		Model llm.Model
		Token string
	}{
		{
			Model: llm.ModelClaude_3_Haiku,
			Token: token,
		},
	}

	for _, test := range tests {
		t.Run(test.Model.String(), func(t *testing.T) {
			c := llm.NewAnthropicClient(llm.NewAnthropicClientOptions{
				Model: test.Model,
				Token: test.Token,
			})

			t.Run("can respond with a message", func(t *testing.T) {
				m := llm.Message{
					Content: "Write me a personalized poem.",
					Name:    "Steve",
					Role:    llm.MessageRoleUser,
				}

				var b strings.Builder
				err := c.Prompt(context.Background(), "You are a disco music and goat fan.", []llm.Message{m}, &b)
				is.NotError(t, err)
				is.True(t, len(b.String()) > 0)
				t.Log(b.String())
			})

			t.Run("can use a system prompt", func(t *testing.T) {
				m := llm.Message{
					Content: "Just say hi.",
					Name:    "Steve",
					Role:    llm.MessageRoleUser,
				}

				var b strings.Builder
				err := c.Prompt(context.Background(), "Always respond in Spanish with just the word 'hola'.", []llm.Message{m}, &b)
				is.NotError(t, err)
				is.True(t, len(b.String()) > 0)
				t.Log(b.String())
				is.True(t, strings.Contains(strings.ToLower(b.String()), "hola"))
			})
		})
	}
}
