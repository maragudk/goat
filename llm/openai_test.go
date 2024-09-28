package llm_test

import (
	"context"
	"strings"
	"testing"

	"github.com/maragudk/is"

	"maragu.dev/goat/llm"
)

func TestOpenAIClient_Prompt(t *testing.T) {
	tests := []struct {
		BaseURL string
		Model   llm.Model
		Token   string
	}{
		{
			BaseURL: "http://localhost:8090/v1",
			Model:   llm.ModelLlama3_2_1B,
		},
	}
	for _, test := range tests {
		t.Run(test.Model.String(), func(t *testing.T) {
			c := llm.NewOpenAIClient(llm.NewOpenAIClientOptions{
				BaseURL: test.BaseURL,
				Model:   test.Model,
				Token:   test.Token,
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

			t.Run("should not include <|eot_id|>", func(t *testing.T) {
				m := llm.Message{
					Content: "Just say hi.",
					Name:    "Steve",
					Role:    llm.MessageRoleUser,
				}

				var b strings.Builder
				err := c.Prompt(context.Background(), "You can only say hi.", []llm.Message{m}, &b)
				is.NotError(t, err)
				is.True(t, len(b.String()) > 0)
				is.True(t, !strings.Contains(b.String(), "<|eot_id|>"))
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
				is.True(t, strings.Contains(b.String(), "Hola."))
			})
		})
	}
}
