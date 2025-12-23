//go:build test
// +build test

package web

import "embed"

// Dummy initialization for tests - webappFS is declared here for tests
// This file ensures webappFS is initialized with an empty embed.FS for tests
// when -tags=test is used
var webappFS embed.FS

