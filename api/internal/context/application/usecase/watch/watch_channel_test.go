package watch

import (
	"slices"
	"testing"

	"repin/internal/context/domain"
)

func TestExpandAlbumRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{
			name: "single id expands both ways",
			in:   []int{15},
			want: rangeInts(5, 25),
		},
		{
			name: "low id is clamped to 1",
			in:   []int{3},
			want: rangeInts(1, 13),
		},
		{
			name: "overlapping ids are deduplicated",
			in:   []int{15, 16},
			want: rangeInts(5, 26),
		},
		{
			name: "scattered ids stay separate ranges",
			in:   []int{15, 100},
			want: append(rangeInts(5, 25), rangeInts(90, 110)...),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := expandAlbumRange(tt.in); !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelatedMessages(t *testing.T) {
	t.Parallel()

	msg := func(id int, groupID int64) domain.RawMessage {
		return domain.RawMessage{ID: id, GroupID: groupID}
	}

	messages := []domain.RawMessage{
		msg(1, 0),
		msg(2, 100),
		msg(3, 100),
		msg(4, 100),
		msg(5, 0),
		msg(6, 200),
		msg(7, 200),
	}

	tests := []struct {
		name string
		ids  []int
		want []int
	}{
		{
			name: "plain message keeps only itself",
			ids:  []int{5},
			want: []int{5},
		},
		{
			name: "album member pulls the whole album",
			ids:  []int{3},
			want: []int{2, 3, 4},
		},
		{
			name: "mixed ids pull albums and singletons",
			ids:  []int{1, 6},
			want: []int{1, 6, 7},
		},
		{
			name: "unknown id yields nothing",
			ids:  []int{99},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := relatedMessages(messages, tt.ids)

			ids := make([]int, 0, len(got))
			for _, m := range got {
				ids = append(ids, m.ID)
			}

			if !slices.Equal(ids, tt.want) {
				t.Errorf("got %v, want %v", ids, tt.want)
			}
		})
	}
}

func rangeInts(from, to int) []int {
	out := make([]int, 0, to-from+1)
	for i := from; i <= to; i++ {
		out = append(out, i)
	}

	return out
}
