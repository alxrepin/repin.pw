package prompts

import (
	"strings"
	"testing"
)

func TestSEOSystemIsTrimmed(t *testing.T) {
	if SEOSystem == "" {
		t.Fatal("SEOSystem is empty")
	}

	if strings.TrimSpace(SEOSystem) != SEOSystem {
		t.Errorf("SEOSystem is not trimmed: %q", SEOSystem)
	}
}
