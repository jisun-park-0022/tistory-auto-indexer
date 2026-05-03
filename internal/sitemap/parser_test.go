package sitemap

import (
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		want    []SitemapURL
		wantErr bool
	}{
		{
			name: "valid sitemap",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>https://example.tistory.com/1</loc><lastmod>2026-05-01</lastmod></url>
  <url><loc>https://example.tistory.com/2</loc><lastmod>2026-05-03</lastmod></url>
</urlset>`,
			want: []SitemapURL{
				{Loc: "https://example.tistory.com/1", LastMod: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)},
				{Loc: "https://example.tistory.com/2", LastMod: time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC)},
			},
		},
		{
			name: "missing lastmod is allowed",
			xml: `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>https://example.tistory.com/3</loc></url>
</urlset>`,
			want: []SitemapURL{
				{Loc: "https://example.tistory.com/3"},
			},
		},
		{
			name:    "invalid XML",
			xml:     `not xml`,
			wantErr: true,
		},
		{
			name: "empty loc is skipped",
			xml: `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc></loc></url>
  <url><loc>https://example.tistory.com/4</loc></url>
</urlset>`,
			want: []SitemapURL{
				{Loc: "https://example.tistory.com/4"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parse(strings.NewReader(tc.xml))
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got.URLs) != len(tc.want) {
				t.Fatalf("got %d URLs, want %d", len(got.URLs), len(tc.want))
			}
			for i, u := range got.URLs {
				if u.Loc != tc.want[i].Loc {
					t.Errorf("[%d] Loc: got %q, want %q", i, u.Loc, tc.want[i].Loc)
				}
				if !u.LastMod.Equal(tc.want[i].LastMod) {
					t.Errorf("[%d] LastMod: got %v, want %v", i, u.LastMod, tc.want[i].LastMod)
				}
			}
		})
	}
}
