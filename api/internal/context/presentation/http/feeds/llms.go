package feeds

import (
	"strings"
	"time"

	"repin/internal/context/domain"
)

func (c *Controller) renderLLMs(src *source) ([]byte, error) {
	var b strings.Builder

	b.WriteString("# " + src.Channel.Title + "\n\n")
	b.WriteString("> " + c.siteDescription(src) + "\n\n")
	b.WriteString("Сайт: " + c.siteURL + "\n")
	b.WriteString("Telegram: https://t.me/" + src.Channel.Name + "\n")
	b.WriteString("Полные тексты всех постов: " + c.siteURL + "/llms-full.txt\n")
	b.WriteString("RSS: " + c.siteURL + "/rss.xml\n\n")
	b.WriteString("## Посты\n\n")

	for i := range src.Posts {
		post := &src.Posts[i]

		b.WriteString("- [" + postTitle(post) + "](" + c.postURL(post) + ")")
		if desc := postDescription(post); desc != "" {
			b.WriteString(": " + desc)
		}

		b.WriteString(" (" + post.CreatedAt.UTC().Format(time.DateOnly) + ")\n")
	}

	return []byte(b.String()), nil
}

func (c *Controller) renderLLMsFull(src *source) ([]byte, error) {
	var b strings.Builder

	b.WriteString("# " + src.Channel.Title + "\n\n")
	b.WriteString("> " + c.siteDescription(src) + "\n\n")
	b.WriteString("Сайт: " + c.siteURL + "\n")
	b.WriteString("Telegram: https://t.me/" + src.Channel.Name + "\n")

	for i := range src.Posts {
		post := &src.Posts[i]

		b.WriteString("\n---\n\n")
		b.WriteString("## " + postTitle(post) + "\n\n")
		b.WriteString("Дата: " + post.CreatedAt.UTC().Format(time.DateOnly) + "\n")
		b.WriteString("URL: " + c.postURL(post) + "\n\n")

		if text := postBody(post); text != "" {
			b.WriteString(text + "\n")
		}
	}

	return []byte(b.String()), nil
}

func postBody(p *domain.Post) string {
	if p.RawText != nil && *p.RawText != "" {
		return strings.TrimSpace(*p.RawText)
	}

	return ""
}
