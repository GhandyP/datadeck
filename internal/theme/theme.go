// Package theme holds the single stylesheet both surfaces are built from, so
// the app shell and the exported deck cannot drift apart.
package theme

import _ "embed"

//go:embed tokens.css
var tokens string

// Tokens returns the shared design tokens.
func Tokens() string { return tokens }
