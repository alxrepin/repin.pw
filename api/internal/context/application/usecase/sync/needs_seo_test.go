package sync

import (
	"testing"

	"repin/internal/context/domain"
)

func TestNeedsSEO(t *testing.T) {
	t.Parallel()

	ptr := func(s string) *string { return &s }

	complete := func() *domain.Post {
		return &domain.Post{
			RawText:        ptr("text"),
			SEOTitle:       ptr("title"),
			SEODescription: ptr("description"),
			SEOKeywords:    ptr("a, b"),
		}
	}

	tests := []struct {
		name     string
		existing *domain.Post
		post     *domain.Post
		want     bool
	}{
		{
			name: "new post with text",
			post: &domain.Post{RawText: ptr("text")},
			want: true,
		},
		{
			name: "new post without text",
			post: &domain.Post{},
			want: false,
		},
		{
			name: "new post with empty text",
			post: &domain.Post{RawText: ptr("")},
			want: false,
		},
		{
			name:     "unchanged text keeps existing metadata",
			existing: complete(),
			post:     &domain.Post{RawText: ptr("text")},
			want:     false,
		},
		{
			name:     "edited text triggers regeneration",
			existing: complete(),
			post:     &domain.Post{RawText: ptr("other text")},
			want:     true,
		},
		{
			name: "partial metadata is regenerated",
			existing: &domain.Post{
				RawText:  ptr("text"),
				SEOTitle: ptr("title"),
			},
			post: &domain.Post{RawText: ptr("text")},
			want: true,
		},
		{
			name:     "post imported before metadata existed",
			existing: &domain.Post{RawText: ptr("text")},
			post:     &domain.Post{RawText: ptr("text")},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := needsSEO(tt.existing, tt.post); got != tt.want {
				t.Errorf("needsSEO() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMediaMessageIDs(t *testing.T) {
	t.Parallel()

	group := []domain.RawMessage{
		{ID: 1},
		{ID: 2, Media: &domain.Media{ID: 20}},
		{ID: 3},
		{ID: 4, Media: &domain.Media{ID: 40}},
	}

	got := mediaMessageIDs(group)
	want := []int64{2, 4}

	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
