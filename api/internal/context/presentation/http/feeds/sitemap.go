package feeds

import (
	"context"
	"encoding/xml"
	"time"

	"github.com/rs/zerolog"

	"repin/internal/context/domain"
)

const (
	sitemapMaxURLs = 50_000
	sitemapBatch   = 1_000
)

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

func (c *Controller) renderSitemap(ctx context.Context) ([]byte, error) {
	set := sitemapURLSet{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9"}

	set.URLs = append(set.URLs, sitemapURL{Loc: c.siteURL + "/"})

	var (
		afterID int64
		latest  string
	)

scan:
	for {
		posts, err := c.posts.ListAfter(ctx, afterID, sitemapBatch)
		if err != nil {
			return nil, err
		}

		if len(posts) == 0 {
			break
		}

		for i := range posts {
			if len(set.URLs) > sitemapMaxURLs {
				zerolog.Ctx(ctx).Warn().
					Int("max", sitemapMaxURLs).
					Msg("sitemap truncated: protocol cap reached, a sitemap index is needed")

				break scan
			}

			mod := lastMod(&posts[i])
			if mod > latest {
				latest = mod
			}

			set.URLs = append(set.URLs, sitemapURL{Loc: c.postURL(&posts[i]), LastMod: mod})
		}

		afterID = posts[len(posts)-1].ID
	}

	set.URLs[0].LastMod = latest

	body, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), body...), nil
}

func lastMod(p *domain.Post) string {
	t := p.CreatedAt
	if p.UpdatedAt != nil && p.UpdatedAt.After(t) {
		t = *p.UpdatedAt
	}

	return t.UTC().Format(time.RFC3339)
}
