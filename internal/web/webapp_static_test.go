//go:build test
// +build test

package web

// Dummy initialization for tests - webappFS is declared in webapp_routes.go
// This file ensures webappFS is initialized with an empty embed.FS for tests
// when -tags=test is used
var _ = webappFS

