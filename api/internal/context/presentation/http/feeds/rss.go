package feeds

import (
	"encoding/xml"
	"strconv"
	"time"

	"repin/internal/context/presentation/http/media"
)

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	AtomLink    atomLink  `xml:"atom:link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	Items       []rssItem `xml:"item"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItem struct {
	Title       string        `xml:"title"`
	Link        string        `xml:"link"`
	GUID        rssGUID       `xml:"guid"`
	PubDate     string        `xml:"pubDate"`
	Description *rssCDATA     `xml:"description,omitempty"`
	Enclosure   *rssEnclosure `xml:"enclosure,omitempty"`
}

type rssGUID struct {
	Value       string `xml:",chardata"`
	IsPermaLink bool   `xml:"isPermaLink,attr"`
}

type rssCDATA struct {
	Value string `xml:",cdata"`
}

type rssEnclosure struct {
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

func (c *Controller) renderRSS(src *source) ([]byte, error) {
	feed := rssFeed{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Channel: rssChannel{
			Title:       src.Channel.Title,
			Link:        c.siteURL + "/",
			AtomLink:    atomLink{Href: c.siteURL + "/rss.xml", Rel: "self", Type: "application/rss+xml"},
			Description: c.siteDescription(src),
			Language:    "ru",
			Items:       make([]rssItem, 0, len(src.Posts)),
		},
	}

	for i := range src.Posts {
		post := &src.Posts[i]

		item := rssItem{
			Title:   postTitle(post),
			Link:    c.postURL(post),
			GUID:    rssGUID{Value: c.postURL(post), IsPermaLink: true},
			PubDate: post.CreatedAt.UTC().Format(time.RFC1123Z),
		}

		if post.Text != nil && *post.Text != "" {
			item.Description = &rssCDATA{Value: *post.Text}
		}

		if photo := firstPhoto(post); photo != nil {
			enc := rssEnclosure{URL: media.URL(c.mediaURL, photo.ObjectKey), Type: "image/jpeg"}
			if photo.MimeType != nil && *photo.MimeType != "" {
				enc.Type = *photo.MimeType
			}

			if photo.Size != nil {
				enc.Length = strconv.FormatInt(*photo.Size, 10)
			}

			item.Enclosure = &enc
		}

		feed.Channel.Items = append(feed.Channel.Items, item)
	}

	body, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), body...), nil
}
