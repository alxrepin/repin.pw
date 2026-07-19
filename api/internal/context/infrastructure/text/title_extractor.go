package text

import (
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

	title := strings.TrimSpace(reTag.ReplaceAllString(strings.TrimSpace(match[1]), ""))
	remaining := strings.TrimSpace(reH1.ReplaceAllString(text, ""))

	return title, remaining
}
