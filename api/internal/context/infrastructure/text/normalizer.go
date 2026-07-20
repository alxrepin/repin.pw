package text

import (
	"html"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"repin/internal/context/domain"
)

var reNewlineBeforeClose = regexp.MustCompile(`\n(</[^>]+>)`)

type span struct {
	start   int
	end     int
	opening string
	closing string
}

// covers reports whether the span covers the whole [from, to) segment.
func (s span) covers(from, to int) bool { return s.start <= from && s.end >= to }

// Normalizer converts Telegram text + formatting entities into safe HTML: the
// text is escaped, hrefs are scheme-checked, and overlapping entities are
// rendered by closing and reopening tags so nesting stays valid.
type Normalizer struct{}

func NewNormalizer() *Normalizer { return &Normalizer{} }

// Normalize renders the message. channel is the username the post came from,
// used to recognise links pointing back into the same channel; empty disables
// that rewriting and every link is then treated as external.
func (n *Normalizer) Normalize(text string, entities []domain.RawMessageEntity, channel string) string {
	if text == "" {
		return ""
	}

	spans := buildSpans(text, entities, channel)

	// Boundaries chop the text into segments, each covered by a fixed set of
	// entities; a segment's tag set is diffed against the previously open one.
	bounds := []int{0, len(text)}
	for _, s := range spans {
		bounds = append(bounds, s.start, s.end)
	}

	sort.Ints(bounds)
	bounds = slices.Compact(bounds)

	var (
		b    strings.Builder
		open []int // indices into spans, in tag-opening order
	)

	for i := 0; i+1 < len(bounds); i++ {
		from, to := bounds[i], bounds[i+1]

		var active []int

		for idx, s := range spans {
			if s.covers(from, to) {
				active = append(active, idx)
			}
		}

		// Keep the common prefix of open tags; close the rest in reverse and
		// open what the segment needs. Reopening is what keeps overlapping
		// entities properly nested.
		common := 0
		for common < len(open) && common < len(active) && open[common] == active[common] {
			common++
		}

		for j := len(open) - 1; j >= common; j-- {
			b.WriteString(spans[open[j]].closing)
		}

		for _, idx := range active[common:] {
			b.WriteString(spans[idx].opening)
		}

		open = active

		b.WriteString(html.EscapeString(text[from:to]))
	}

	for j := len(open) - 1; j >= 0; j-- {
		b.WriteString(spans[open[j]].closing)
	}

	return reNewlineBeforeClose.ReplaceAllString(b.String(), "$1\n")
}

func buildSpans(text string, entities []domain.RawMessageEntity, channel string) []span {
	spans := make([]span, 0, len(entities))

	for _, e := range entities {
		start := utf16ToByteOffset(text, e.Offset)
		end := utf16ToByteOffset(text, e.Offset+e.Length)

		if start >= end || end > len(text) {
			continue
		}

		opening, closing, ok := tagFor(e, text[start:end], channel)
		if !ok {
			continue
		}

		spans = append(spans, span{start: start, end: end, opening: opening, closing: closing})
	}

	sort.SliceStable(spans, func(i, j int) bool {
		if spans[i].start != spans[j].start {
			return spans[i].start < spans[j].start
		}

		return spans[i].end > spans[j].end
	})

	return spans
}

func tagFor(e domain.RawMessageEntity, content string, channel string) (opening, closing string, ok bool) {
	switch e.Type {
	case domain.EntityTypeBold:
		return "<strong>", "</strong>", true
	case domain.EntityTypeItalic:
		return "<em>", "</em>", true
	case domain.EntityTypeUnderline:
		return "<u>", "</u>", true
	case domain.EntityTypeStrike:
		return "<s>", "</s>", true
	case domain.EntityTypeSpoiler:
		return `<span class="spoiler">`, "</span>", true
	case domain.EntityTypeBlockquote:
		if e.Collapsed {
			return `<blockquote data-collapsed>`, "</blockquote>", true
		}

		return "<blockquote>", "</blockquote>", true
	case domain.EntityTypeCode:
		return "<code>", "</code>", true
	case domain.EntityTypePre:
		if e.Language != nil && *e.Language != "" {
			return `<pre><code class="language-` + html.EscapeString(*e.Language) + `">`, "</code></pre>", true
		}

		return "<pre>", "</pre>", true
	case domain.EntityTypeTextLink, domain.EntityTypeURL:
		raw := content
		if e.URL != nil {
			raw = *e.URL
		}

		href, valid := safeHref(raw)
		if !valid {
			return "", "", false
		}

		// A link back into our own channel is a link to a page of this site, so
		// it keeps the tab and gets handled by the router. Everything else
		// leaves, and leaving opens a new tab. rel is not decoration there:
		// without noopener the opened page can reach back through window.opener.
		if internal, ok := internalPostHref(href, channel); ok {
			return `<a href="` + html.EscapeString(internal) + `">`, "</a>", true
		}

		return `<a href="` + html.EscapeString(href) + `" target="_blank" rel="noopener noreferrer">`, "</a>", true
	case domain.EntityTypeCustomEmoji:
		if e.CustomEmojiID != nil {
			return `<span data-emoji-id="` + strconv.FormatInt(*e.CustomEmojiID, 10) + `">`, "</span>", true
		}
	}

	return "", "", false
}

var telegramHosts = map[string]bool{"t.me": true, "telegram.me": true, "telegram.dog": true}

// internalPostHref turns a link to a post of our own channel into a link to the
// same post on this site. Telegram numbers the messages of a channel, and that
// number is the post id here, so "t.me/<channel>/42" and "/posts/42" name the
// same thing. Links to any other channel are left alone — they are external.
func internalPostHref(href, channel string) (string, bool) {
	if channel == "" {
		return "", false
	}

	u, err := url.Parse(href)
	if err != nil || !telegramHosts[strings.ToLower(u.Hostname())] {
		return "", false
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")

	// "/s/<channel>/<n>" is the web-preview form of the same link.
	if len(parts) == 3 && strings.EqualFold(parts[0], "s") {
		parts = parts[1:]
	}

	if len(parts) != 2 || !strings.EqualFold(parts[0], channel) {
		return "", false
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return "", false
	}

	return "/posts/" + strconv.FormatInt(id, 10), true
}

// safeHref validates a link target, allowing only benign schemes; scheme-less
// targets (Telegram links plain "example.com") default to https.
func safeHref(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)

	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}

	switch strings.ToLower(u.Scheme) {
	case "http", "https", "mailto", "tg":
		return raw, true
	case "":
		return "https://" + raw, true
	default:
		return "", false
	}
}
