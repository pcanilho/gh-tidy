package main

import (
	"github.com/pcanilho/gh-tidy/cmd"
	"log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
