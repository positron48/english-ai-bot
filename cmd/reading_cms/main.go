package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"tgbot-skeleton/internal/readingcms"
)

func main() {
	port := flag.Int("port", 8791, "HTTP port")
	repoRoot := flag.String("repo", "", "repository root (default: auto-detect)")
	flag.Parse()

	root := *repoRoot
	if root == "" {
		wd, _ := os.Getwd()
		detected, err := readingcms.FindRepoRoot(wd)
		if err != nil {
			log.Fatal(err)
		}
		root = detected
	}

	svc, err := readingcms.NewService(root)
	if err != nil {
		log.Fatal(err)
	}
	webRoot := filepath.Join(root, "tools", "reading-cms", "web")
	srv := readingcms.NewServer(svc, webRoot)

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	addr := ln.Addr().String()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Reading Texts CMS (local)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Open: http://%s/\n", addr)
	fmt.Printf("Data: %s\n", svc.Paths().DataDir)
	fmt.Println("Stop: Ctrl+C")

	if err := http.Serve(ln, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
