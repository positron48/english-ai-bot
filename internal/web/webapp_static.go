//go:build !test
// +build !test

package web

import "embed"

//go:embed webapp/dist
var _ = webappFS // Initialize webappFS with embedded files in production

