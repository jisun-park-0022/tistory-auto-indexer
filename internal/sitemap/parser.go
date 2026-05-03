package sitemap

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Parser interface {
	Fetch(ctx context.Context, sitemapURL string) (*Sitemap, error)
}

type HTTPParser struct {
	client    *http.Client
	userAgent string
}

func NewHTTPParser(timeout time.Duration, userAgent string) *HTTPParser {
	return &HTTPParser{
		client:    &http.Client{Timeout: timeout},
		userAgent: userAgent,
	}
}

func (p *HTTPParser) Fetch(ctx context.Context, sitemapURL string) (*Sitemap, error) {
	slog.Debug("fetching sitemap", "url", sitemapURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sitemapURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", p.userAgent)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching sitemap %s: %w", sitemapURL, err)
	}
	defer resp.Body.Close()

	slog.Debug("sitemap response received", "status", resp.StatusCode, "url", sitemapURL)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching sitemap", resp.StatusCode)
	}

	sm, err := parse(resp.Body)
	if err != nil {
		return nil, err
	}

	slog.Info("sitemap parsed", "url", sitemapURL, "url_count", len(sm.URLs))
	return sm, nil
}

// xmlURLSet mirrors the standard Sitemap 0.9 XML structure.
type xmlURLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []xmlURL `xml:"url"`
}

type xmlURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

func parse(r io.Reader) (*Sitemap, error) {
	var raw xmlURLSet
	if err := xml.NewDecoder(r).Decode(&raw); err != nil {
		return nil, fmt.Errorf("parsing sitemap XML: %w", err)
	}

	slog.Debug("sitemap XML decoded", "raw_url_count", len(raw.URLs))

	sm := &Sitemap{URLs: make([]SitemapURL, 0, len(raw.URLs))}
	for _, u := range raw.URLs {
		if u.Loc == "" {
			slog.Debug("skipping entry with empty loc")
			continue
		}
		su := SitemapURL{Loc: u.Loc}
		if u.LastMod != "" {
			t, err := time.Parse("2006-01-02", u.LastMod)
			if err == nil {
				su.LastMod = t
			} else {
				slog.Debug("failed to parse lastmod, ignoring", "loc", u.Loc, "lastmod", u.LastMod)
			}
		}
		sm.URLs = append(sm.URLs, su)
	}
	return sm, nil
}
