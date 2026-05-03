package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jisun/tistory-indexer/internal/gsc"
	"github.com/jisun/tistory-indexer/internal/sitemap"
	"github.com/jisun/tistory-indexer/internal/state"
	"github.com/jisun/tistory-indexer/pkg/config"
)

type Service struct {
	parser sitemap.Parser   //interface
	gsc    gsc.GSCClient    //interface
	store  state.StateStore //interface
	cfg    *config.Config
}

func NewService(
	parser sitemap.Parser,
	gscClient gsc.GSCClient,
	store state.StateStore,
	cfg *config.Config,
) *Service {
	return &Service{
		parser: parser,
		gsc:    gscClient,
		store:  store,
		cfg:    cfg,
	}
}

func (s *Service) Run(ctx context.Context) error {
	st, err := s.store.Load()
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	sm, err := s.parser.Fetch(ctx, s.cfg.Tistory.SitemapURL)
	if err != nil {
		return fmt.Errorf("fetching sitemap: %w", err)
	}

	currentURLs := sm.Locs()
	now := time.Now().UTC()

	// first run: no prior state — check GSC before deciding whether to submit
	if st.LastRunAt.IsZero() {
		st.LastRunAt = now
		return s.initializeState(ctx, st, currentURLs, now)
	}

	st.LastRunAt = now

	newURLs := detectNewURLs(currentURLs, st.KnownURLs)
	if len(newURLs) == 0 {
		slog.Info("no new posts detected, skipping GSC submission")
		if err := s.store.Save(st); err != nil {
			return fmt.Errorf("saving state: %w", err)
		}
		return nil
	}

	slog.Info("new posts detected, submitting sitemap to GSC", "count", len(newURLs))
	for _, u := range newURLs {
		slog.Info("new post", "url", u)
	}

	if err := s.gsc.SubmitSitemap(ctx, s.cfg.Google.SiteURL, s.cfg.Google.SitemapURL); err != nil {
		slog.Error("GSC submission failed, state not updated", "err", err)
		return fmt.Errorf("submitting sitemap to GSC: %w", err)
	}

	st.KnownURLs = currentURLs
	st.LastSubmittedAt = now

	if err := s.store.Save(st); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	slog.Info("sitemap submitted and state updated", "submitted_at", st.LastSubmittedAt)
	return nil
}

// initializeState handles the very first run.
// If the sitemap is already registered in GSC, treat current URLs as the baseline
// and skip submission. Otherwise, submit and save.
func (s *Service) initializeState(ctx context.Context, st *state.State, currentURLs []string, now time.Time) error {
	already, err := s.gsc.HasSitemap(ctx, s.cfg.Google.SiteURL, s.cfg.Google.SitemapURL)
	if err != nil {
		return fmt.Errorf("checking GSC sitemap registration: %w", err)
	}

	st.KnownURLs = currentURLs

	if already {
		slog.Info("first run: sitemap already registered in GSC, saving baseline without submission",
			"url_count", len(currentURLs))
	} else {
		slog.Info("first run: sitemap not found in GSC, submitting", "url_count", len(currentURLs))
		if err := s.gsc.SubmitSitemap(ctx, s.cfg.Google.SiteURL, s.cfg.Google.SitemapURL); err != nil {
			return fmt.Errorf("submitting sitemap to GSC: %w", err)
		}
		st.LastSubmittedAt = now
	}

	if err := s.store.Save(st); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}
	return nil
}

// detectNewURLs returns URLs in current that are not in known. O(n).
func detectNewURLs(current, known []string) []string {
	knownSet := make(map[string]struct{}, len(known))
	for _, u := range known {
		knownSet[u] = struct{}{}
	}

	var newURLs []string
	for _, u := range current {
		if _, exists := knownSet[u]; !exists {
			newURLs = append(newURLs, u)
		}
	}
	return newURLs
}
