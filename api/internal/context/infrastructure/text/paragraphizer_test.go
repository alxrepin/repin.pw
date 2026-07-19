package text

import "testing"

func TestParagraphize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty",
			in:   "",
			want: "",
		},
		{
			name: "single newline becomes br",
			in:   "one\ntwo",
			want: "<p>one<br>two</p>",
		},
		{
			name: "blank line splits paragraphs",
			in:   "one\n\ntwo",
			want: "<p>one</p>\n<p>two</p>",
		},
		{
			name: "bullet emoji lines become a list",
			in:   "intro:\n\n🔵 first\n🔵 second",
			want: "<p>intro:</p>\n<ul><li>first</li><li>second</li></ul>",
		},
		{
			name: "blank lines between bullets keep one list",
			in:   "🔵 first\n\n🔵 second\n\nafter",
			want: "<ul><li>first</li><li>second</li></ul>\n<p>after</p>",
		},
		{
			name: "custom emoji bullet is stripped",
			in:   `<span data-emoji-id="5400228162703484613">🔵</span>Публикация стоит 5$`,
			want: "<ul><li>Публикация стоит 5$</li></ul>",
		},
		{
			name: "bold-wrapped custom emoji bullet is stripped",
			in:   `<strong><span data-emoji-id="5215584915898243758">🟢</span></strong>В стране есть <strong>дети</strong>`,
			want: "<ul><li>В стране есть <strong>дети</strong></li></ul>",
		},
		{
			name: "line after bullet continues the item",
			in:   "🔹 first line\nsecond line",
			want: "<ul><li>first line<br>second line</li></ul>",
		},
		{
			name: "headings pass through as blocks",
			in:   "<h2><strong>Head</strong></h2>\ntext",
			want: "<h2><strong>Head</strong></h2>\n<p>text</p>",
		},
		{
			name: "bold bullet title becomes a list item, not a heading",
			in:   `<strong><span data-emoji-id="1">🟢</span></strong><strong>Рисовые поля</strong>` + "\n\ntext",
			want: "<ul><li><strong>Рисовые поля</strong></li></ul>\n<p>text</p>",
		},
		{
			name: "pre keeps its newlines",
			in:   "before\n<pre>a\nb</pre>\nafter",
			want: "<p>before</p>\n<pre>a\nb</pre>\n<p>after</p>",
		},
		{
			name: "blockquote newlines become br",
			in:   "<blockquote>line1\nline2</blockquote>\ntail",
			want: "<blockquote>line1<br>line2</blockquote>\n<p>tail</p>",
		},
		{
			name: "lone bullet emoji line is dropped",
			in:   "one\n🔵\ntwo",
			want: "<p>one</p>\n<p>two</p>",
		},
	}

	p := NewParagraphizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := p.Format(tt.in); got != tt.want {
				t.Errorf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExcerpt(t *testing.T) {
	t.Parallel()

	got := Excerpt("<p>one</p><ul><li>two</li></ul>", 100)
	if got != "one two" {
		t.Errorf("Excerpt() = %q, want %q", got, "one two")
	}

	if got := Excerpt("<p>абвгд</p>", 3); got != "абв" {
		t.Errorf("Excerpt() truncated = %q, want %q", got, "абв")
	}
}
