package main

import (
	"bufio"
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"maragu.dev/errors"
	"maragu.dev/migrate"
	"maragu.dev/snorkel"

	"maragu.dev/goat/llm"
	"maragu.dev/goo/sql"
)

//go:embed sql/migrations
var migrations embed.FS

func main() {
	log := snorkel.New(snorkel.Options{})

	if err := start(log); err != nil {
		log.Event("Error starting", 1, "error", err)
	}
}

func start(log *snorkel.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Make a .goat directory in the user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "error getting home directory")
	}
	goatDir := filepath.Join(home, ".goat")
	if err := os.MkdirAll(goatDir, 0700); err != nil {
		return errors.Wrap(err, "error creating .goat directory")
	}

	h := sql.NewHelper(sql.NewHelperOptions{
		Path: filepath.Join(goatDir, "goat.db"),
	})
	if err := h.Connect(); err != nil {
		return errors.Wrap(err, "error connecting to database")
	}

	subFS, err := fs.Sub(migrations, "sql/migrations")
	if err != nil {
		return errors.Wrap(err, "error getting sub filesystem")
	}

	if err := migrate.Up(ctx, h.DB.DB, subFS); err != nil {
		return errors.Wrap(err, "error migrating")
	}

	var messages []llm.Message
	c := llm.NewOpenAIClient(llm.NewOpenAIClientOptions{
		BaseURL: "http://localhost:8090/v1",
		Model:   llm.ModelLlama3_2_1B,
	})

	fmt.Print("> ")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()

		fmt.Println()

		messages = append(messages, llm.Message{
			Content: text,
			Name:    "Me",
			Role:    llm.MessageRoleUser,
		})

		var b strings.Builder
		w := io.MultiWriter(os.Stdout, &b)

		if err := c.Prompt(ctx, "You are an assistant.", messages, w); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Error: ", err)
		}

		fmt.Println()
		fmt.Println()

		messages = append(messages, llm.Message{
			Content: b.String(),
			Name:    "Assistant",
			Role:    llm.MessageRoleAssistant,
		})

		fmt.Print("> ")
	}

	return nil
}
