package main

import (
	"fmt"
	"github.com/pcanilho/gh-tidy/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprint(os.Stderr, err, "\n")
		os.Exit(1)
	}
}
