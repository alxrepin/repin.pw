package prompts

import (
	_ "embed"
	"strings"
)

//go:embed seo_system.md
var seoSystem string

var SEOSystem = strings.TrimSpace(seoSystem)
