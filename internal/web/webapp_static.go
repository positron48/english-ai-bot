//go:build !test
// +build !test

package web

import "embed"

//go:embed ../../webapp/dist
var webappFS embed.FS

