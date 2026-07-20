package text

import (
	"regexp"
	"strings"

	"github.com/essentialkaos/translit"
)

var (
	reSpaces  = regexp.MustCompile(`[\s\-_]+`)
	reInvalid = regexp.MustCompile(`[^a-z0-9\-_]`)
	reDashes  = regexp.MustCompile(`-+`)
)

func Slug(text string) string {
	if text == "" {
		return ""
	}

	s := strings.ToLower(translit.EncodeToICAO(text))
	s = reSpaces.ReplaceAllString(s, "-")
	s = reInvalid.ReplaceAllString(s, "")
	s = strings.Trim(s, "-")
	s = reDashes.ReplaceAllString(s, "-")

	return s
}

// TruncateSlug cuts a slug down to maxLen bytes at a word boundary. Slugs are
// ASCII, so byte length equals character length.
func TruncateSlug(slug string, maxLen int) string {
	if len(slug) <= maxLen {
		return slug
	}

	cut := slug[:maxLen]
	if i := strings.LastIndex(cut, "-"); i > 0 {
		cut = cut[:i]
	}

	return strings.Trim(cut, "-")
}

func utf16ToByteOffset(text string, utf16Pos int) int {
	var bytePos, current int

	for _, r := range text {
		if current == utf16Pos {
			return bytePos
		}

		inc := 1
		if r > 0xFFFF {
			inc = 2
		}

		current += inc
		bytePos += len(string(r))
	}

	return bytePos
}
