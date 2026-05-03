package app

import (
	"context"
	"fmt"

	"github.com/jisun/tistory-indexer/internal/gsc"
	"github.com/jisun/tistory-indexer/internal/indexer"
	"github.com/jisun/tistory-indexer/internal/sitemap"
	"github.com/jisun/tistory-indexer/internal/state"
	"github.com/jisun/tistory-indexer/pkg/config"
)

type App struct {
	Service *indexer.Service
	Store   state.StateStore
	Config  *config.Config
}

func Build(ctx context.Context) (*App, error) {
	cfg, err := config.Load(".env", "config.yaml")
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	gscClient, err := gsc.NewClient(ctx, cfg.Google.ClientID, cfg.Google.ClientSecret, cfg.Google.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("creating GSC client: %w", err)
	}

	parser := sitemap.NewHTTPParser(cfg.HTTP.Timeout(), cfg.HTTP.UserAgent)
	store := state.NewFileStore(cfg.State.FilePath)
	svc := indexer.NewService(parser, gscClient, store, cfg)

	return &App{
		Service: svc,
		Store:   store,
		Config:  cfg,
	}, nil
}
