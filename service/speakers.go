package service

import (
	"fmt"

	"maragu.dev/clir"
	"maragu.dev/errors"
)

func (s *Service) PrintSpeakers(ctx clir.Context) error {
	speakers, err := s.DB.GetSpeakers(ctx.Ctx)
	if err != nil {
		return errors.Wrap(err, "error getting speakers")
	}
	for _, s := range speakers {
		_, _ = fmt.Fprintf(ctx.Out, "- %v (%v)\n", s.Name, s.ModelName)
		if s.System != "" {
			_, _ = fmt.Fprintf(ctx.Out, "  system: \"%v\"\n", s.System)
		}
	}
	return nil
}
