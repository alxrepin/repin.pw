package text

import (
	"testing"

	"repin/internal/context/domain"
)

func TestNormalize(t *testing.T) {
	t.Parallel()

	str := func(s string) *string { return &s }
	id := func(v int64) *int64 { return &v }

	ent := func(typ domain.RawMessageEntityType, offset, length int) domain.RawMessageEntity {
		return domain.RawMessageEntity{Type: typ, Offset: offset, Length: length}
	}

	tests := []struct {
		name     string
		text     string
		entities []domain.RawMessageEntity
		want     string
	}{
		{
			name: "plain text is escaped even without entities",
			text: `a < b & "c"`,
			want: "a &lt; b &amp; &#34;c&#34;",
		},
		{
			name:     "bold",
			text:     "hello world",
			entities: []domain.RawMessageEntity{ent(domain.EntityTypeBold, 0, 5)},
			want:     "<strong>hello</strong> world",
		},
		{
			name: "new entity types",
			text: "abcd",
			entities: []domain.RawMessageEntity{
				ent(domain.EntityTypeUnderline, 0, 1),
				ent(domain.EntityTypeStrike, 1, 1),
				ent(domain.EntityTypeSpoiler, 2, 1),
				ent(domain.EntityTypeBlockquote, 3, 1),
			},
			want: `<u>a</u><s>b</s><span class="spoiler">c</span><blockquote>d</blockquote>`,
		},
		{
			name: "collapsed blockquote",
			text: "quote",
			entities: []domain.RawMessageEntity{
				{Type: domain.EntityTypeBlockquote, Offset: 0, Length: 5, Collapsed: true},
			},
			want: "<blockquote data-collapsed>quote</blockquote>",
		},
		{
			name: "pre with language",
			text: "x := 1",
			entities: []domain.RawMessageEntity{
				{Type: domain.EntityTypePre, Offset: 0, Length: 6, Language: str("go")},
			},
			want: `<pre><code class="language-go">x := 1</code></pre>`,
		},
		{
			name:     "escaped text inside entity",
			text:     "a<b>c",
			entities: []domain.RawMessageEntity{ent(domain.EntityTypeCode, 0, 5)},
			want:     "<code>a&lt;b&gt;c</code>",
		},
		{
			name: "text link with quote in url is escaped",
			text: "link",
			entities: []domain.RawMessageEntity{
				{Type: domain.EntityTypeTextLink, Offset: 0, Length: 4, URL: str(`https://x.com/?q="><script>`)},
			},
			want: `<a href="https://x.com/?q=&#34;&gt;&lt;script&gt;" target="_blank" rel="noopener noreferrer">link</a>`,
		},
		{
			name: "javascript scheme is dropped",
			text: "link",
			entities: []domain.RawMessageEntity{
				{Type: domain.EntityTypeTextLink, Offset: 0, Length: 4, URL: str("javascript:alert(1)")},
			},
			want: "link",
		},
		{
			name:     "bare url is linked with https fallback",
			text:     "see example.com now",
			entities: []domain.RawMessageEntity{ent(domain.EntityTypeURL, 4, 11)},
			want:     `see <a href="https://example.com" target="_blank" rel="noopener noreferrer">example.com</a> now`,
		},
		{
			name: "overlapping entities stay properly nested",
			text: "abcdefghij",
			entities: []domain.RawMessageEntity{
				ent(domain.EntityTypeBold, 0, 6),
				ent(domain.EntityTypeItalic, 3, 7),
			},
			want: "<strong>abc<em>def</em></strong><em>ghij</em>",
		},
		{
			name: "nested entities open outer first",
			text: "abcdef",
			entities: []domain.RawMessageEntity{
				ent(domain.EntityTypeItalic, 2, 2),
				ent(domain.EntityTypeBold, 0, 6),
			},
			want: "<strong>ab<em>cd</em>ef</strong>",
		},
		{
			name: "utf16 offsets with surrogate pairs",
			text: "🔥 hot",
			entities: []domain.RawMessageEntity{
				ent(domain.EntityTypeBold, 3, 3), // "🔥" is 2 UTF-16 units
			},
			want: "🔥 <strong>hot</strong>",
		},
		{
			name: "custom emoji",
			text: "😀",
			entities: []domain.RawMessageEntity{
				{Type: domain.EntityTypeCustomEmoji, Offset: 0, Length: 2, CustomEmojiID: id(42)},
			},
			want: `<span data-emoji-id="42">😀</span>`,
		},
		{
			name: "trailing newline moves out of the closing tag",
			text: "title\nbody",
			entities: []domain.RawMessageEntity{
				ent(domain.EntityTypeBold, 0, 6),
			},
			want: "<strong>title</strong>\nbody",
		},
		{
			name: "out of range entity is dropped",
			text: "short",
			entities: []domain.RawMessageEntity{
				ent(domain.EntityTypeBold, 3, 100),
			},
			want: "sho<strong>rt</strong>",
		},
	}

	n := NewNormalizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := n.Normalize(tt.text, tt.entities, ""); got != tt.want {
				t.Errorf("Normalize() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRendererTitleAndBody(t *testing.T) {
	t.Parallel()

	r := NewRenderer()

	title, body := r.Render("Заголовок. И сразу текст\nвторая строка", nil, "")

	if title != "Заголовок." {
		t.Errorf("title = %q, want %q", title, "Заголовок.")
	}

	// The first line's tail must survive title extraction, not be dropped.
	if body != "<p>И сразу текст<br>вторая строка</p>" {
		t.Errorf("body = %q, want %q", body, "<p>И сразу текст<br>вторая строка</p>")
	}
}

func TestRendererBlockquoteFirstLineIsNotTitle(t *testing.T) {
	t.Parallel()

	r := NewRenderer()

	title, _ := r.Render("цитата\n\nтекст", []domain.RawMessageEntity{
		{Type: domain.EntityTypeBlockquote, Offset: 0, Length: 6},
	}, "")

	if title != "" {
		t.Errorf("title = %q, want empty: a quote is not a title", title)
	}
}
