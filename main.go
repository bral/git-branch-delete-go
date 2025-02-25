package main

import (
	"os"

	"github.com/bral/git-branch-delete-go/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
