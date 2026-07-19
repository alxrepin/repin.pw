package openrouter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"repin/internal/context/domain"
)

const maxInputRunes = 6000

const seoSystemPrompt = `Ты SEO-редактор блога. По тексту поста ты составляешь метаданные для поисковой выдачи.

Правила:
- Пиши на том же языке, что и пост.
- seo_title: до 60 символов, отражает суть поста, без кликбейта и без названия сайта.
- seo_description: 120-160 символов, законченное предложение, передаёт пользу поста для читателя.
- seo_keywords: 5-8 ключевых фраз в нижнем регистре, по убыванию релевантности, без повторов и без решёток.
- Опирайся только на текст поста, ничего не выдумывай.`

const (
	jsType = "type"
	jsDesc = "description"
)

func stringProp(description string) map[string]any {
	return map[string]any{jsType: "string", jsDesc: description}
}

var seoSchema = Schema{
	Name: "post_seo",
	Definition: map[string]any{
		jsType: "object",
		"properties": map[string]any{
			"seo_title":       stringProp("Заголовок для поисковой выдачи, до 60 символов"),
			"seo_description": stringProp("Описание для сниппета, 120-160 символов"),
			"seo_keywords": map[string]any{
				jsType:  "array",
				jsDesc:  "5-8 ключевых фраз",
				"items": map[string]any{jsType: "string"},
			},
		},
		"required":             []string{"seo_title", "seo_description", "seo_keywords"},
		"additionalProperties": false,
	},
}

type seoResult struct {
	Title       string   `json:"seo_title"`
	Description string   `json:"seo_description"`
	Keywords    []string `json:"seo_keywords"`
}

type SEOGenerator struct {
	client *Client
}

func NewSEOGenerator(client *Client) *SEOGenerator {
	return &SEOGenerator{client: client}
}

var ErrNothingToDescribe = errors.New("post has no text to describe")

func (g *SEOGenerator) Generate(ctx context.Context, post *domain.Post) (*domain.PostSEO, error) {
	source := postText(post)
	if source == "" {
		return nil, ErrNothingToDescribe
	}

	messages := []Message{
		{Role: "system", Content: seoSystemPrompt},
		{Role: "user", Content: source},
	}

	var result seoResult
	if err := g.client.CompleteJSON(ctx, messages, seoSchema, &result); err != nil {
		return nil, err
	}

	seo := &domain.PostSEO{
		Title:       strings.TrimSpace(result.Title),
		Description: strings.TrimSpace(result.Description),
		Keywords:    joinKeywords(result.Keywords),
	}

	if seo.Title == "" || seo.Description == "" {
		return nil, fmt.Errorf("openrouter: incomplete seo for post %d", post.ID)
	}

	return seo, nil
}

func postText(post *domain.Post) string {
	var b strings.Builder

	if post.Title != nil && *post.Title != "" {
		b.WriteString(*post.Title)
		b.WriteString("\n\n")
	}

	if post.RawText != nil {
		b.WriteString(*post.RawText)
	}

	return truncate(strings.TrimSpace(b.String()), maxInputRunes)
}

func truncate(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}

	return string(runes[:limit])
}

func joinKeywords(keywords []string) string {
	out := make([]string, 0, len(keywords))
	seen := make(map[string]struct{}, len(keywords))

	for _, k := range keywords {
		k = strings.ToLower(strings.TrimSpace(k))
		if k == "" {
			continue
		}

		if _, dup := seen[k]; dup {
			continue
		}

		seen[k] = struct{}{}

		out = append(out, k)
	}

	return strings.Join(out, ", ")
}
