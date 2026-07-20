package service

import "testing"

// The canonical slug is "<id>-<title>", so the leading number alone has to
// identify the post: that is what makes internal Telegram links and stale
// slugs resolve.
func TestLeadingID(t *testing.T) {
	tests := []struct {
		slug string
		want int64
		ok   bool
	}{
		{"42-nazvanie-posta", 42, true},
		{"42", 42, true},
		{"86-gt-v-kitae-nichego-ne-poniatno", 86, true},
		{"42-", 42, true},
		{"", 0, false},
		{"nazvanie-bez-id", 0, false},
		{"-42", 0, false},
		{"0", 0, false},
		{"0-nol-ne-id", 0, false},
		{"-1", 0, false},
		{"4.2-drob", 0, false},
		{"99999999999999999999-perepolnenie", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.slug, func(t *testing.T) {
			got, ok := leadingID(tt.slug)
			if ok != tt.ok || got != tt.want {
				t.Errorf("leadingID(%q) = (%d, %v), want (%d, %v)", tt.slug, got, ok, tt.want, tt.ok)
			}
		})
	}
}
