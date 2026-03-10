package main

import (
	"fmt"
	"os"

	appcli "github.com/felipeelias/claude-statusline/internal/cli"
)

var version = "dev"

func main() {
	app := appcli.New(version)

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
