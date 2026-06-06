package main

import (
	"fmt"
	"os"

	"github.com/guferreira1/spec-harbor/internal/adapters/cli"
)

func main() {
	if err := cli.Execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
