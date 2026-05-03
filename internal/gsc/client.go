package gsc

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/webmasters/v3"
)

const (
	scope      = "https://www.googleapis.com/auth/webmasters"
	maxRetries = 3
)

type GSCClient interface {
	SubmitSitemap(ctx context.Context, siteURL, sitemapURL string) error
	HasSitemap(ctx context.Context, siteURL, sitemapURL string) (bool, error)
}

type Client struct {
	svc *webmasters.Service
}

func NewClient(ctx context.Context, clientID, clientSecret, refreshToken string) (*Client, error) {
	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{scope},
	}

	token := &oauth2.Token{RefreshToken: refreshToken}
	httpClient := cfg.Client(ctx, token)

	svc, err := webmasters.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("creating webmasters service: %w", err)
	}

	return &Client{svc: svc}, nil
}

func (c *Client) SubmitSitemap(ctx context.Context, siteURL, sitemapURL string) error {
	slog.Info("submitting sitemap to GSC", "site_url", siteURL, "sitemap_url", sitemapURL)

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			slog.Warn("rate limited, retrying", "attempt", attempt, "backoff_sec", backoff.Seconds())
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		err := c.submit(ctx, siteURL, sitemapURL)
		if err == nil {
			slog.Info("sitemap submitted successfully", "site_url", siteURL, "sitemap_url", sitemapURL)
			return nil
		}

		// retry only on rate limit (429)
		if isRateLimit(err) {
			lastErr = err
			continue
		}
		return err
	}
	return fmt.Errorf("submit failed after %d retries: %w", maxRetries, lastErr)
}

func (c *Client) submit(ctx context.Context, siteURL, sitemapURL string) error {
	call := c.svc.Sitemaps.Submit(siteURL, sitemapURL)
	if err := call.Context(ctx).Do(); err != nil {
		slog.Error("failed to submit sitemap", "site_url", siteURL, "sitemap_url", sitemapURL, "error", err)
		return fmt.Errorf("submitting sitemap: %w", err)
	}
	slog.Info("sitemap submitted successfully", "site_url", siteURL, "sitemap_url", sitemapURL)
	return nil
}

func (c *Client) HasSitemap(ctx context.Context, siteURL, sitemapURL string) (bool, error) {
	slog.Debug("checking sitemap registration in GSC", "site_url", siteURL, "sitemap_url", sitemapURL)

	resp, err := c.svc.Sitemaps.List(siteURL).Context(ctx).Do()
	if err != nil {
		return false, fmt.Errorf("listing sitemaps: %w", err)
	}

	slog.Debug("GSC sitemaps listed", "count", len(resp.Sitemap))

	found := false
	for _, s := range resp.Sitemap {
		slog.Debug("GSC registered sitemap",
			"path", s.Path,
			"type", s.Type,
			"last_submitted", s.LastSubmitted,
			"last_downloaded", s.LastDownloaded,
			"is_pending", s.IsPending,
			"warnings", s.Warnings,
			"errors", s.Errors,
		)
		if s.Path == sitemapURL {
			found = true
		}
	}
	return found, nil
	return false, nil
}

func isRateLimit(err error) bool {
	if apiErr, ok := err.(interface{ Code() int }); ok {
		return apiErr.Code() == http.StatusTooManyRequests
	}
	return false
}
