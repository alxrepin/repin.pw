package text

import "strings"

// HeaderAnalyzer promotes the leading line of normalized text to an <h1> and
// standalone bold lines to <h2>, so a title can later be extracted.
type HeaderAnalyzer struct{}

func NewHeaderAnalyzer() *HeaderAnalyzer { return &HeaderAnalyzer{} }

// Analyze injects <h1>/<h2> heading tags into text.
func (a *HeaderAnalyzer) Analyze(text string) string {
	if text == "" {
		return ""
	}

	lines := strings.Split(text, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// A post opening with block-level content, or a bullet-marked line
		// (which the paragraphizer turns into a list item), has no title.
		if strings.HasPrefix(trimmed, "<blockquote") || strings.HasPrefix(trimmed, "<pre") ||
			reBullet.MatchString(trimmed) {
			break
		}

		content, rest := line, ""

		if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "" {
			// Cut at the first sentence, but never through markup: a dot past
			// a tag may sit inside an attribute or split a tag pair.
			if dot := strings.Index(content, "."); dot != -1 && !strings.Contains(content[:dot], "<") {
				content, rest = content[:dot+1], strings.TrimSpace(content[dot+1:])
			}
		}

		lines[i] = "<h1>" + content + "</h1>"
		if rest != "" {
			lines[i] += "\n" + rest
		}

		break
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if !strings.HasPrefix(trimmed, "<strong>") || !strings.HasSuffix(trimmed, "</strong>") {
			continue
		}

		// A bullet-marked bold line is a list item, not a subheading: leave it
		// for the paragraphizer.
		if reBullet.MatchString(trimmed) {
			continue
		}

		prevEmpty := i == 0 || strings.TrimSpace(lines[i-1]) == ""
		nextEmptyOrNotBold := i+1 >= len(lines) ||
			strings.TrimSpace(lines[i+1]) == "" ||
			!strings.HasPrefix(strings.TrimSpace(lines[i+1]), "<strong>")

		if prevEmpty && nextEmptyOrNotBold {
			lines[i] = "<h2>" + trimmed + "</h2>"
		}
	}

	return strings.Join(lines, "\n")
}
