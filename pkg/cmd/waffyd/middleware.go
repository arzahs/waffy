package waffyd

import (
	"gopkg.in/urfave/cli.v1"

	"github.com/unerror/waffy/pkg/config"
	"github.com/unerror/waffy/pkg/data"
)

func withConfig(f func(ctx *cli.Context, cfg *config.Config) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return f(ctx, cfg)
	}
}

func withDatabase(f func(ctx *cli.Context, s data.Store) error) func(*cli.Context) error {
	return withConfig(func(ctx *cli.Context, cfg *config.Config) error {
		db, err := data.NewDB(cfg.DBPath)
		if err != nil {
			return err
		}

		return f(ctx, db)
	})
}

func withDatabaseConfig(f func(ctx *cli.Context, s data.Store, cfg *config.Config) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		var c *config.Config
		var db data.Store

		// TODO: better way to do this config
		err := withConfig(func(ctx *cli.Context, cfg *config.Config) error {
			c = cfg
			return nil
		})(ctx)
		if err != nil {
			return err
		}

		err = withDatabase(func(ctx *cli.Context, s data.Store) error {
			db = s
			return nil
		})(ctx)
		if err != nil {
			return err
		}

		return f(ctx, db, c)
	}
}

func withConsensus(f func(ctx *cli.Context, c data.Consensus) error) func(*cli.Context) error {
	return withDatabaseConfig(func(ctx *cli.Context, s data.Store, cfg *config.Config) error {
		raft, err := data.NewRaft(cfg.RaftDIR, cfg.RaftListen, s)
		if err != nil {
			return err
		}

		return f(ctx, raft)
	})
}
