package service

import (
	"encoding/json"
	"fmt"

	"maragu.dev/clir"
	"maragu.dev/errors"

	"maragu.dev/goat/model"
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

func (s *Service) AddModel(ctx clir.Context, modelName, modelType, token, address string) error {
	if modelName == "" {
		return errors.New("model name is required")
	}
	if modelType == "" {
		return errors.New("model type is required")
	}

	config := map[string]string{}
	m := model.Model{
		Name: modelName,
		Type: model.ModelType(modelType),
	}
	if token != "" {
		config["token"] = token
	}
	if address != "" {
		config["address"] = address
	}
	m.Config = mustMarshalJSON(config)

	if err := s.DB.SaveModel(ctx.Ctx, m); err != nil {
		return err
	}
	return nil
}

func mustMarshalJSON(config map[string]string) string {
	b, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	return string(b)
}
