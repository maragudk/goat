package main

import (
	"embed"
	"flag"
	"os"
	"path/filepath"

	"maragu.dev/clir"
	"maragu.dev/clir/middleware"
	"maragu.dev/env"
	"maragu.dev/errors"

	"maragu.dev/goat/service"
)

//go:embed public
var public embed.FS

func main() {
	clir.Run(clir.RunnerFunc(func(ctx clir.Context) error {
		_ = env.Load()

		dir := env.GetStringOrDefault("GOAT_DIR", "")
		if dir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return errors.Wrap(err, "error getting home directory")
			}
			dir = filepath.Join(home, ".goat")
		}
		if err := os.MkdirAll(dir, 0700); err != nil {
			return errors.Wrap(err, "error creating .goat directory")
		}

		s := service.New(service.NewOptions{
			Path: dir,
		})

		if err := s.ConnectAndMigrate(ctx.Ctx); err != nil {
			return err
		}

		r := clir.NewRouter()

		var continueFlag *bool
		var promptFlag *string
		r.Use(middleware.Flags(func(fs *flag.FlagSet) {
			continueFlag = fs.Bool("c", false, "continue conversation")
			promptFlag = fs.String("p", "", "use a one-off prompt instead of chatting")
		}))

		r.RouteFunc("", func(ctx clir.Context) error {
			opts := service.StartOptions{
				Continue: *continueFlag,
				Prompt:   *promptFlag,
			}
			return s.Start(ctx.Ctx, ctx.In, ctx.Out, opts)
		})

		r.Branch("models", func(r *clir.Router) {
			r.RouteFunc("", s.PrintModels)
			r.RouteFunc("list", s.PrintModels)
		})

		r.Branch("speakers", func(r *clir.Router) {
			r.RouteFunc("", s.PrintSpeakers)
			r.RouteFunc("list", s.PrintSpeakers)
		})

		r.Branch("conversations", func(r *clir.Router) {
			r.RouteFunc("", s.PrintConversations)
			r.RouteFunc("list", s.PrintConversations)
			r.RouteFunc("recompute-topics", s.RecomputeTopics)
		})

		r.RouteFunc("serve", func(ctx clir.Context) error {
			if err := os.Setenv("DATABASE_PATH", filepath.Join(dir, "goat.db")); err != nil {
				return errors.Wrap(err, "error setting DATABASE_PATH")
			}
			if err := os.Setenv("SERVER_ADDRESS", ":9090"); err != nil {
				return errors.Wrap(err, "error setting SERVER_ADDRESS")
			}
			s.Serve(ctx.Ctx, s.DB, public, ctx.Err)
			return nil
		})

		return r.Run(ctx)
	}))
}
