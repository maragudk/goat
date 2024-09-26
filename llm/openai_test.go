package llm_test

import (
	"context"
	"strings"
	"testing"

	"github.com/maragudk/is"

	"maragu.dev/goat/llm"
)

func TestOpenAIClient_Prompt(t *testing.T) {
	t.Run("can respond with a message", func(t *testing.T) {
		tests := []struct {
			BaseURL string
			Model   llm.Model
			Skip    bool
			Token   string
		}{
			{
				BaseURL: "http://localhost:8091/v1",
				Model:   llm.ModelLlama3_2_1B,
				Skip:    false,
			},
			{
				BaseURL: "http://localhost:8092/v1",
				Model:   llm.ModelLlama3_2_3B,
				Skip:    true,
			},
		}

		for _, test := range tests {
			t.Run(test.Model.String(), func(t *testing.T) {
				if test.Skip {
					t.Skip()
				}

				c := llm.NewOpenAIClient(llm.NewOpenAIClientOptions{
					BaseURL: test.BaseURL,
					Model:   test.Model,
					Token:   test.Token,
				})

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
		}
	})
}
