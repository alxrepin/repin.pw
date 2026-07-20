package text

import (
	"strings"
	"testing"

	"repin/internal/context/domain"
)

func TestTitleExtractorDecodesEntities(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"greater than", "<h1>&gt; в Китае ничего не понятно</h1><p>x</p>", "> в Китае ничего не понятно"},
		{"ampersand", "<h1>Rock &amp; Roll</h1>", "Rock & Roll"},
		{"quotes", "<h1>&#34;цитата&#34; и &#39;апостроф&#39;</h1>", `"цитата" и 'апостроф'`},
		{"escaped tag stays text", "<h1>&lt;b&gt;не тег&lt;/b&gt;</h1>", "<b>не тег</b>"},
		{"real tags stripped", "<h1>жирный <b>кусок</b></h1>", "жирный кусок"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := NewTitleExtractor().Extract(tt.in)
			if got != tt.want {
				t.Errorf("Extract() title = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExcerptDecodesEntities(t *testing.T) {
	got := Excerpt("<p>&gt; там не работают карты<br>&gt; без языка никак</p>", 100)
	want := "> там не работают карты > без языка никак"
	if got != want {
		t.Errorf("Excerpt() = %q, want %q", got, want)
	}
}

func TestExcerptTruncatesAfterDecoding(t *testing.T) {
	got := Excerpt("<p>&gt;&gt;&gt;&gt;&gt;abc</p>", 5)
	if want := ">>>>>"; got != want {
		t.Errorf("Excerpt() = %q, want %q", got, want)
	}
	if strings.Contains(got, "&") {
		t.Errorf("Excerpt() leaked an entity: %q", got)
	}
}

func TestRendererProducesPlainTextTitle(t *testing.T) {
	raw := "> в Китае ничего не понятно\n\n> там не работают карты"
	title, body := NewRenderer().Render(raw, []domain.RawMessageEntity{})

	if strings.Contains(title, "&") {
		t.Errorf("title still carries an entity: %q", title)
	}
	if !strings.HasPrefix(title, ">") {
		t.Errorf("title = %q, want it to start with %q", title, ">")
	}

	if !strings.Contains(body, "&gt; там не работают карты") {
		t.Errorf("body should keep its escaping, got %q", body)
	}
}
