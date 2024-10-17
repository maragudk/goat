package service

import (
	"fmt"
	"strings"

	"maragu.dev/clir"
	"maragu.dev/errors"

	"maragu.dev/goat/llm"
	"maragu.dev/goat/model"
)

func (s *Service) RecomputeTopics(ctx clir.Context) error {
	out := ctx.Out

	models, err := s.DB.GetModels(ctx.Ctx)
	if err != nil {
		return errors.Wrap(err, "error getting models")
	}

	var client prompter
	for _, m := range models {
		if m.Type != model.ModelTypeOpenAI {
			continue
		}

		_, _ = fmt.Fprintln(out, "Using model:", m.Name)
		client = newClientFromModel(m)
		break
	}

	if client == nil {
		return errors.New("no models available to recompute topics")
	}

	cds, err := s.DB.GetConversationDocuments(ctx.Ctx)
	if err != nil {
		return errors.Wrap(err, "error getting conversation documents")
	}
	for _, cd := range cds {
		var messages []llm.Message
		for _, t := range cd.Turns {
			speaker := cd.Speakers[t.SpeakerID]

			messages = append(messages, llm.Message{
				Content: speaker.Name + ": " + t.Content,
				Name:    speaker.Name,
				Role:    llm.MessageRoleUser,
			})
		}

		var b strings.Builder
		if err := client.Prompt(ctx.Ctx, model.CreateSummarizerPrompt("Summarizer"), messages, &b); err != nil {
			return errors.Wrap(err, "error summarizing conversation")
		}

		if err := s.DB.SaveTopic(ctx.Ctx, cd.Conversation.ID, b.String()); err != nil {
			return errors.Wrap(err, "error saving conversation topic")
		}

		_, _ = fmt.Fprintf(out, `Recomputed topic for conversation %v: "%v"\n`, cd.Conversation.ID, b.String())
	}

	return nil
}
