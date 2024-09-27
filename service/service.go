package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"maragu.dev/errors"

	"maragu.dev/goat/llm"
	goosql "maragu.dev/goo/sql"

	"maragu.dev/goat/sql"
)

type Service struct {
	DB *sql.Database
}

type NewOptions struct {
	Path string
}

func New(opts NewOptions) *Service {
	h := goosql.NewHelper(goosql.NewHelperOptions{
		Path: filepath.Join(opts.Path, "goat.db"),
	})
	db := sql.NewDatabase(sql.NewDatabaseOptions{
		SQLHelper: h,
	})

	return &Service{
		DB: db,
	}
}

func (s *Service) Start(ctx context.Context, r io.Reader, w io.Writer) error {
	if err := s.DB.Connect(); err != nil {
		return errors.Wrap(err, "error connecting to database")
	}

	if err := s.DB.MigrateUp(ctx); err != nil {
		return errors.Wrap(err, "error migrating database")
	}

	var messages []llm.Message
	c := llm.NewOpenAIClient(llm.NewOpenAIClientOptions{
		BaseURL: "http://localhost:8090/v1",
		Model:   llm.ModelLlama3_2_1B,
	})

	_, _ = fmt.Fprint(w, "> ")

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()

		_, _ = fmt.Fprintln(w)

		messages = append(messages, llm.Message{
			Content: text,
			Name:    "Me",
			Role:    llm.MessageRoleUser,
		})

		var b strings.Builder
		multiW := io.MultiWriter(w, &b)

		if err := c.Prompt(ctx, "You are an assistant.", messages, multiW); err != nil {
			_, _ = fmt.Fprintln(multiW, "\n\nError: ", err)
		}

		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintln(w)

		messages = append(messages, llm.Message{
			Content: b.String(),
			Name:    "Assistant",
			Role:    llm.MessageRoleAssistant,
		})

		fmt.Print("> ")
	}
	return nil
}
