package cmd

import (
	"gopkg.in/urfave/cli.v1"

	"github.com/unerror/waffy/pkg/config"
	"github.com/unerror/waffy/pkg/data"
)

// WithConfig will pass the local configuration object as well as the context to the Action
func WithConfig(f func(ctx *cli.Context, cfg *config.Config) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		return f(ctx, cfg)
	}
}

// WithDatabase will pass the local datastore as well as the context to the Action
func WithDatabase(f func(ctx *cli.Context, s data.Store) error) func(*cli.Context) error {
	return WithConfig(func(ctx *cli.Context, cfg *config.Config) error {
		db, err := data.NewDB(cfg.DBPath)
		if err != nil {
			return err
		}

		return f(ctx, db)
	})
}

// WithDatabaseConfig will pass the local datastore, the local coniguration as well as the context
// to the Action
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

// WithConsensus will pass the consensus data store as well as the context to the Action
func WithConsensus(f func(ctx *cli.Context, c data.Consensus) error) func(*cli.Context) error {
	return WithDatabaseConfig(func(ctx *cli.Context, s data.Store, cfg *config.Config) error {
		raft, err := data.NewRaft(cfg.RaftDIR, cfg.RaftListen, s)
		if err != nil {
			return err
		}

		return f(ctx, raft)
	})
}

// WithConsensusConfig will pass the consensus data store, and the local configuration as well as
// the context to the Action
func WithConsensusConfig(f func(ctx *cli.Context, s data.Consensus, c *config.Config) error) func(*cli.Context) error {
	return func(ctx *cli.Context) error {
		var c *config.Config
		var s data.Consensus

		// TODO: better way to do this config too
		err := WithConfig(func(ctx *cli.Context, cfg *config.Config) error {
			c = cfg
			return nil
		})(ctx)
		if err != nil {
			return err
		}

		err = WithConsensus(func(ctx *cli.Context, c data.Consensus) error {
			s = c
			return nil
		})(ctx)
		if err != nil {
			return err
		}

		return f(ctx, s, c)
	}
}
