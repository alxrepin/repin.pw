package sync

import (
	"testing"

	"repin/internal/context/domain"
)

func TestGroupMessages(t *testing.T) {
	t.Parallel()

	msg := func(id int, groupID int64) domain.RawMessage {
		return domain.RawMessage{ID: id, GroupID: groupID}
	}

	tests := []struct {
		name string
		in   []domain.RawMessage
		want [][]int // expected message IDs per group, in order
	}{
		{
			name: "empty",
			in:   nil,
			want: nil,
		},
		{
			name: "singletons keep ascending order",
			in:   []domain.RawMessage{msg(3, 0), msg(1, 0), msg(2, 0)},
			want: [][]int{{1}, {2}, {3}},
		},
		{
			name: "album is merged and sorted by id",
			in:   []domain.RawMessage{msg(7, 100), msg(5, 100), msg(6, 100)},
			want: [][]int{{5, 6, 7}},
		},
		{
			name: "albums and singletons interleaved newest-first",
			in: []domain.RawMessage{
				msg(9, 0),
				msg(8, 200),
				msg(7, 200),
				msg(4, 0),
				msg(3, 100),
				msg(2, 100),
			},
			want: [][]int{{2, 3}, {4}, {7, 8}, {9}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := groupMessages(tt.in)

			if len(got) != len(tt.want) {
				t.Fatalf("got %d groups, want %d", len(got), len(tt.want))
			}

			for i, group := range got {
				if len(group) != len(tt.want[i]) {
					t.Fatalf("group %d: got %d messages, want %d", i, len(group), len(tt.want[i]))
				}

				for j, m := range group {
					if m.ID != tt.want[i][j] {
						t.Errorf("group %d message %d: got id %d, want %d", i, j, m.ID, tt.want[i][j])
					}
				}
			}
		})
	}
}
