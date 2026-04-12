package main

import (
	"fmt"
	"os"
)

func main() {
	os.Exit(run())
}

func run() int {
	fmt.Println("verb_form_examples: cloze training now uses runtime Spanish/Russian pairs (see internal/spanishverbs/dynamic_example_pair.go).")
	fmt.Println("No database rows are required for verb cloze cards; this command exits successfully without changes.")
	return 0
}
