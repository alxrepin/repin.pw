package text

import (
	"testing"

	"repin/internal/context/domain"
)

func link(url string) []domain.RawMessageEntity {
	return []domain.RawMessageEntity{{Type: domain.EntityTypeTextLink, Offset: 0, Length: 3, URL: &url}}
}

func TestInternalLinksBecomeSiteLinks(t *testing.T) {
	tests := []struct {
		name    string
		href    string
		channel string
		want    string
	}{
		{
			name:    "own channel post",
			href:    "https://t.me/allrpn/42",
			channel: "allrpn",
			want:    `<a href="/posts/42">тут</a>`,
		},
		{
			name:    "channel match is case insensitive",
			href:    "https://t.me/AllRPN/42",
			channel: "allrpn",
			want:    `<a href="/posts/42">тут</a>`,
		},
		{
			name:    "web preview form",
			href:    "https://t.me/s/allrpn/7",
			channel: "allrpn",
			want:    `<a href="/posts/7">тут</a>`,
		},
		{
			name:    "telegram.me mirror",
			href:    "https://telegram.me/allrpn/7",
			channel: "allrpn",
			want:    `<a href="/posts/7">тут</a>`,
		},
		{
			name:    "another channel stays external",
			href:    "https://t.me/someoneelse/42",
			channel: "allrpn",
			want:    `<a href="https://t.me/someoneelse/42" target="_blank" rel="noopener noreferrer">тут</a>`,
		},
		{
			name:    "channel root is not a post",
			href:    "https://t.me/allrpn",
			channel: "allrpn",
			want:    `<a href="https://t.me/allrpn" target="_blank" rel="noopener noreferrer">тут</a>`,
		},
		{
			name:    "non numeric tail is not a post",
			href:    "https://t.me/allrpn/about",
			channel: "allrpn",
			want:    `<a href="https://t.me/allrpn/about" target="_blank" rel="noopener noreferrer">тут</a>`,
		},
		{
			name:    "unknown channel disables rewriting",
			href:    "https://t.me/allrpn/42",
			channel: "",
			want:    `<a href="https://t.me/allrpn/42" target="_blank" rel="noopener noreferrer">тут</a>`,
		},
		{
			name:    "ordinary external link",
			href:    "https://example.com/a",
			channel: "allrpn",
			want:    `<a href="https://example.com/a" target="_blank" rel="noopener noreferrer">тут</a>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewNormalizer().Normalize("тут", link(tt.href), tt.channel)
			if got != tt.want {
				t.Errorf("Normalize()\n got %s\nwant %s", got, tt.want)
			}
		})
	}
}
