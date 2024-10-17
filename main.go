package main

import (
	"embed"
	"flag"
	"os"
	"path/filepath"

	"maragu.dev/clir"
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

		// TODO move the flags inside the route once the router supports it
		flagSet := flag.NewFlagSet("goat", flag.ExitOnError)
		flagSet.SetOutput(ctx.Err)

		continueFlag := flagSet.Bool("c", false, "continue conversation")
		promptFlag := flagSet.String("p", "", "use a one-off prompt instead of chatting")

		_ = flagSet.Parse(ctx.Args)

		opts := service.StartOptions{
			Continue: *continueFlag,
			Prompt:   *promptFlag,
		}

		r := clir.NewRouter()

		r.Route("", start(s, opts))

		r.Branch("models", func(r *clir.Router) {
			r.RouteFunc("", s.PrintModels)
			r.RouteFunc("list", s.PrintModels)
		})

		r.Branch("speakers", func(r *clir.Router) {
			r.RouteFunc("", s.PrintSpeakers)
			r.RouteFunc("list", s.PrintSpeakers)
		})

		r.Branch("conversations", func(r *clir.Router) {
			r.RouteFunc("recompute-topics", s.RecomputeTopics)
		})

		r.Route("serve", serve(s, dir))

		ctx.Args = flagSet.Args()

		return r.Run(ctx)
	}))
}

func start(s *service.Service, opts service.StartOptions) clir.RunnerFunc {
	return func(ctx clir.Context) error {
		return s.Start(ctx.Ctx, ctx.In, ctx.Out, opts)
	}
}

func serve(s *service.Service, dir string) clir.RunnerFunc {
	return func(ctx clir.Context) error {
		if err := os.Setenv("DATABASE_PATH", filepath.Join(dir, "goat.db")); err != nil {
			return errors.Wrap(err, "error setting DATABASE_PATH")
		}
		s.Serve(ctx.Ctx, s.DB, public, ctx.Err)
		return nil
	}
}
