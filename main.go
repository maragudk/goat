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

		var opts service.StartOptions
		r.Use(middleware.Flags(func(fs *flag.FlagSet) {
			fs.BoolVar(&opts.Continue, "c", false, "continue conversation")
			fs.StringVar(&opts.Prompt, "p", "", "use a one-off prompt instead of chatting")
		}))

		r.RouteFunc("", func(ctx clir.Context) error {
			return s.Start(ctx.Ctx, ctx.In, ctx.Out, opts)
		})

		r.Branch("models", func(r *clir.Router) {
			r.RouteFunc("", s.PrintModels)
			r.RouteFunc("list", s.PrintModels)

			r.Branch("add", func(r *clir.Router) {
				var token, address string

				r.Use(middleware.Flags(func(fs *flag.FlagSet) {
					fs.Usage = func() {
						ctx.Errorln("Usage: goat models add -token <token> -address <address> <model name>:<model type>")
					}
					fs.StringVar(&token, "token", "", "model auth token")
					fs.StringVar(&address, "address", "", "model HTTP address")
				}))

				r.RouteFunc(`(\w+):(\w+)`, func(ctx clir.Context) error {
					modelName := ctx.Matches[1]
					modelType := ctx.Matches[2]
					ctx.Println("Adding model", modelName, "of type", modelType, "with token", token, "and address", address)
					return s.AddModel(ctx, modelName, modelType, token, address)
				})
			})
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
