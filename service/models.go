package service

import (
	"fmt"

	"maragu.dev/clir"
	"maragu.dev/errors"
)

func (s *Service) PrintModels(ctx clir.Context) error {
	models, err := s.DB.GetModels(ctx.Ctx)
	if err != nil {
		return errors.Wrap(err, "error getting models")
	}
	for _, m := range models {
		_, _ = fmt.Fprintf(ctx.Out, "- %v (%v)\n", m.Name, m.Type)
	}
	return nil
}
