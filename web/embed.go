package web

import "embed"

// Files contains all embedded HTML templates and static assets (CSS, JS, images).
//
//go:embed templates static
var Files embed.FS
