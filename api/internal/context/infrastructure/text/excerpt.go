package text

import (
	"regexp"
	"strings"
)

var (
	reAnyTag = regexp.MustCompile(`<[^>]*>`)
	reSpace  = regexp.MustCompile(`\s+`)
)

// Excerpt turns rendered post HTML into a plain-text snippet of at most n
// runes: tags become spaces (so adjacent blocks don't glue together), then
// whitespace is collapsed. Truncating the HTML directly would risk cutting
// through a tag.
func Excerpt(html string, n int) string {
	s := reAnyTag.ReplaceAllString(html, " ")
	s = strings.TrimSpace(reSpace.ReplaceAllString(s, " "))

	runes := []rune(s)
	if len(runes) > n {
		return strings.TrimSpace(string(runes[:n]))
	}

	return s
}
