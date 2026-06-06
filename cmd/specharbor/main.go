package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/guferreira1/spec-harbor/internal/adapters/cli"
)

func main() {
	if err := cli.Execute(os.Args[1:]); err != nil {
		var exitError cli.ExitError
		if errors.As(err, &exitError) {
			os.Exit(exitError.Code)
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
