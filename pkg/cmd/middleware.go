package cmd

import (
	"gopkg.in/urfave/cli.v1"

	"github.com/unerror/waffy/pkg/config"
	"github.com/unerror/waffy/pkg/data"
)

func WithConfig(f func(ctx *cli.Context, cfg *config.Config) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return f(ctx, cfg)
	}
}

func WithDatabase(f func(ctx *cli.Context, s data.Store) error) func(*cli.Context) error {
	return WithConfig(func(ctx *cli.Context, cfg *config.Config) error {
		db, err := data.NewDB(cfg.DBPath)
		if err != nil {
			return err
		}

		return f(ctx, db)
	})
}

func WithDatabaseConfig(f func(ctx *cli.Context, s data.Store, cfg *config.Config) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		var c *config.Config
		var db data.Store

		// TODO: better way to do this config
		err := WithConfig(func(ctx *cli.Context, cfg *config.Config) error {
			c = cfg
			return nil
		})(ctx)
		if err != nil {
			return err
		}

		err = WithDatabase(func(ctx *cli.Context, s data.Store) error {
			db = s
			return nil
		})(ctx)
		if err != nil {
			return err
		}

		return f(ctx, db, c)
	}
}

func WithConsensus(f func(ctx *cli.Context, c data.Consensus) error) func(*cli.Context) error {
	return WithDatabaseConfig(func(ctx *cli.Context, s data.Store, cfg *config.Config) error {
		raft, err := data.NewRaft(cfg.RaftDIR, cfg.RaftListen, s)
		if err != nil {
			return err
		}

		return f(ctx, raft)
	})
}

func withConsensusConfig(f func(ctx *cli.Context, s data.Consensus, c *config.Config) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		var c *config.Config
		var s data.Consensus

		// TODO: better way to do this config too
		err := withConfig(func(ctx *cli.Context, cfg *config.Config) error {
			c = cfg
			return nil
		})(ctx)
		if err != nil {
			return err
		}

		err = withConsensus(func(ctx *cli.Context, c data.Consensus) error {
			s = c
			return nil
		})(ctx)
		if err != nil {
			return err
		}

		return f(ctx, s, c)
	}
}
