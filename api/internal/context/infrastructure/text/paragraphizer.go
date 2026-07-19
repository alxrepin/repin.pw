package text

import (
	"regexp"
	"strings"
)

// reBullet matches a "decorative bullet" at the start of a line: a colored
// shape emoji Telegram authors use as a list marker, optionally wrapped in a
// custom-emoji span and/or inline formatting tags (a bold bullet renders as
// <strong><span…>🟢</span></strong>). Such lines become real <li> items.
var reBullet = regexp.MustCompile(`^(?:<(?:strong|em|u|s)>)*` +
	`(?:<span data-emoji-id="\d+">)?` +
	`(?:🔴|🟠|🟡|🟢|🔵|🟣|🟤|⚫|⚪|🟥|🟧|🟨|🟩|🟦|🟪|🟫|⬛|⬜|🔶|🔷|🔸|🔹|▪|▫|◾|◽|•|‣)` +
	`\x{FE0F}?(?:</span>)?(?:</(?:strong|em|u|s)>)*[ \t]*`)

// Paragraphizer converts the line-oriented normalized text into block-level
// HTML: blank-line-separated runs become <p> (single newlines turn into <br>),
// bullet-emoji lines become <ul><li>, and <pre>/<blockquote> chunks pass
// through untouched. Without it the newlines collapse in the browser and the
// whole post renders as one run-on paragraph.
type Paragraphizer struct{}

func NewParagraphizer() *Paragraphizer { return &Paragraphizer{} }

func (p *Paragraphizer) Format(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}

	var (
		out   []string // finished blocks
		para  []string // pending paragraph lines
		list  []string // finished <li> contents of the open list
		item  []string // lines of the <li> being built
		block []string // lines of an open <pre>/<blockquote> chunk
		depth int      // open pre/blockquote tags within the chunk
	)

	flushItem := func() {
		if item != nil {
			list = append(list, strings.Join(item, "<br>"))
			item = nil
		}
	}

	flushList := func() {
		flushItem()

		if len(list) > 0 {
			var b strings.Builder

			b.WriteString("<ul>")

			for _, li := range list {
				b.WriteString("<li>" + li + "</li>")
			}

			b.WriteString("</ul>")
			out = append(out, b.String())
			list = nil
		}
	}

	flushPara := func() {
		if len(para) > 0 {
			out = append(out, "<p>"+strings.Join(para, "<br>")+"</p>")
			para = nil
		}
	}

	flushBlock := func() {
		// Newlines are preserved inside <pre> (pre-formatted) and become
		// visible breaks inside <blockquote>.
		joined := strings.Join(block, "\n")
		if !strings.HasPrefix(strings.TrimSpace(block[0]), "<pre") {
			joined = strings.Join(block, "<br>")
		}

		out = append(out, joined)
		block = nil
	}

	for line := range strings.SplitSeq(s, "\n") {
		if block != nil {
			block = append(block, line)

			depth += openDelta(line)
			if depth == 0 {
				flushBlock()
			}

			continue
		}

		trimmed := strings.TrimSpace(line)

		if d := openDelta(line); d > 0 || strings.Contains(line, "<pre") || strings.Contains(line, "<blockquote") {
			flushPara()
			flushList()

			if d > 0 {
				block = []string{line}
				depth = d
			} else {
				out = append(out, line) // self-contained on one line
			}

			continue
		}

		switch {
		case trimmed == "":
			// Ends the paragraph and the current list item, but not the list:
			// authors often separate bullet lines with blank lines.
			flushPara()
			flushItem()

		case strings.HasPrefix(trimmed, "<h1") || strings.HasPrefix(trimmed, "<h2"):
			flushPara()
			flushList()
			out = append(out, trimmed)

		case reBullet.MatchString(trimmed):
			flushPara()
			flushItem()

			if content := reBullet.ReplaceAllString(trimmed, ""); content != "" {
				item = []string{content}
			}

		case item != nil:
			// Unbroken line right after a bullet continues that item.
			item = append(item, trimmed)

		default:
			flushList()
			para = append(para, trimmed)
		}
	}

	if block != nil {
		flushBlock()
	}

	flushPara()
	flushList()

	return strings.Join(out, "\n")
}

// openDelta is the change in pre/blockquote nesting depth a line introduces.
func openDelta(line string) int {
	return strings.Count(line, "<pre") + strings.Count(line, "<blockquote") -
		strings.Count(line, "</pre>") - strings.Count(line, "</blockquote>")
}
