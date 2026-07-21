package sync

import (
	"strings"
	"testing"
	"time"

	"repin/internal/context/domain"
	"repin/internal/context/infrastructure/text"
)

func testSync() *SyncChannel {
	return &SyncChannel{renderer: text.NewRenderer()}
}

func str(s string) *string { return &s }

func photo() *domain.Media {
	return &domain.Media{Type: domain.MediaTypePhoto, ID: 1}
}

func TestBuildPostSkipsUnpublishable(t *testing.T) {
	t.Parallel()

	// No title and no media: a bullet-list opening never becomes a title,
	// so this text is not worth a page.
	group := []domain.RawMessage{{ID: 10, Text: str("🔴 пункт списка\n🔴 ещё пункт"), Date: time.Now()}}

	if post := testSync().buildPost(10, group, ""); post != nil {
		t.Fatalf("buildPost() = %+v, want nil", post)
	}
}

func TestBuildPostMediaOnly(t *testing.T) {
	t.Parallel()

	group := []domain.RawMessage{{ID: 10, Media: photo(), Date: time.Now()}}

	post := testSync().buildPost(10, group, "")
	if post == nil {
		t.Fatal("buildPost() = nil, want a post: media alone is publishable")
	}

	if post.Title != nil {
		t.Errorf("Title = %q, want nil", *post.Title)
	}

	if post.URL == nil || *post.URL != "10" {
		t.Errorf("URL = %v, want %q", post.URL, "10")
	}
}

func TestBuildPostTitleAndSlug(t *testing.T) {
	t.Parallel()

	group := []domain.RawMessage{{ID: 42, Text: str("Заголовок. И сразу текст\nвторая строка"), Date: time.Now()}}

	post := testSync().buildPost(42, group, "")
	if post == nil {
		t.Fatal("buildPost() = nil, want a post")
	}

	if post.Title == nil || *post.Title != "Заголовок." {
		t.Errorf("Title = %v, want %q", post.Title, "Заголовок.")
	}

	if post.URL == nil || *post.URL != "42-zagolovok" {
		t.Errorf("URL = %v, want %q", post.URL, "42-zagolovok")
	}

	if post.Text == nil || !strings.Contains(*post.Text, "И сразу текст") {
		t.Errorf("Text = %v, want rendered body", post.Text)
	}
}

func TestBuildPostLongTitleSlugIsTruncated(t *testing.T) {
	t.Parallel()

	title := strings.Repeat("слово ", 40) // way past maxSlugLen once transliterated
	group := []domain.RawMessage{{ID: 42, Text: str(title + "\n\nтело"), Date: time.Now()}}

	post := testSync().buildPost(42, group, "")
	if post == nil {
		t.Fatal("buildPost() = nil, want a post")
	}

	slug := strings.TrimPrefix(*post.URL, "42-")
	if len(slug) > maxSlugLen {
		t.Errorf("slug length = %d, want <= %d", len(slug), maxSlugLen)
	}

	if strings.HasSuffix(slug, "-") {
		t.Errorf("slug %q ends with a dash", slug)
	}
}

func TestBuildPostAlbumPicksLongestCaption(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	group := []domain.RawMessage{
		{ID: 10, GroupID: 77, Media: photo(), Text: str("Кратко"), Date: date},
		{ID: 11, GroupID: 77, Media: photo(), Text: str("Заголовок альбома\n\nДлинная подпись со всеми деталями"), Date: date.Add(time.Second)},
	}

	post := testSync().buildPost(10, group, "")
	if post == nil {
		t.Fatal("buildPost() = nil, want a post")
	}

	if post.RawText == nil || !strings.HasPrefix(*post.RawText, "Заголовок альбома") {
		t.Errorf("RawText = %v, want the longest caption", post.RawText)
	}

	if post.GroupID != 77 {
		t.Errorf("GroupID = %d, want 77", post.GroupID)
	}

	if !post.CreatedAt.Equal(date) {
		t.Errorf("CreatedAt = %v, want the first message's date %v", post.CreatedAt, date)
	}
}

func TestBuildPostInvertMediaPropagates(t *testing.T) {
	t.Parallel()

	group := []domain.RawMessage{
		{ID: 10, GroupID: 77, Media: photo(), Date: time.Now()},
		{ID: 11, GroupID: 77, Media: photo(), InvertMedia: true, Date: time.Now()},
	}

	post := testSync().buildPost(10, group, "")
	if post == nil {
		t.Fatal("buildPost() = nil, want a post")
	}

	if !post.InvertMedia {
		t.Error("InvertMedia = false, want true when any album member has it")
	}
}
