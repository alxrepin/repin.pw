package text

import "repin/internal/context/domain"

type Renderer struct {
	normalizer *Normalizer
	headers    *HeaderAnalyzer
	titles     *TitleExtractor
	paragraphs *Paragraphizer
}

func NewRenderer() *Renderer {
	return &Renderer{
		normalizer: NewNormalizer(),
		headers:    NewHeaderAnalyzer(),
		titles:     NewTitleExtractor(),
		paragraphs: NewParagraphizer(),
	}
}

func (r *Renderer) Render(raw string, entities []domain.RawMessageEntity) (title, body string) {
	normalized := r.normalizer.Normalize(raw, entities)
	analyzed := r.headers.Analyze(normalized)
	title, body = r.titles.Extract(analyzed)

	return title, r.paragraphs.Format(body)
}
