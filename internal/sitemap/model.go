package sitemap

import "time"

type Sitemap struct {
	URLs []SitemapURL
}

type SitemapURL struct {
	Loc     string
	LastMod time.Time
}

func (s *Sitemap) Locs() []string {
	locs := make([]string, len(s.URLs))
	for i, u := range s.URLs {
		locs[i] = u.Loc
	}
	return locs
}
