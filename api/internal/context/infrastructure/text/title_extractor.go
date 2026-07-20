package text

import (
	"html"
	"regexp"
	"strings"
)

var (
	reH1  = regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`)
	reTag = regexp.MustCompile(`<[^>]*>`)
)

type TitleExtractor struct{}

func NewTitleExtractor() *TitleExtractor { return &TitleExtractor{} }

func (e *TitleExtractor) Extract(text string) (string, string) {
	if text == "" {
		return "", ""
	}

	match := reH1.FindStringSubmatch(text)
	if len(match) < 2 {
		return "", text
	}

	// The h1 body is HTML, but the title is stored and rendered as plain text,
	// so entities have to be decoded — otherwise a ">" typed in Telegram reaches
	// the page as "&gt;". Decoding after the tags are stripped, never before:
	// the other order would let an escaped "<b>" turn into a tag and vanish.
	title := strings.TrimSpace(reTag.ReplaceAllString(strings.TrimSpace(match[1]), ""))
	title = html.UnescapeString(title)

	remaining := strings.TrimSpace(reH1.ReplaceAllString(text, ""))

	return title, remaining
}
