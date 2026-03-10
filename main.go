package main

import (
	"fmt"
	"os"

	"github.com/felipeelias/claude-statusline/internal/config"
	"github.com/felipeelias/claude-statusline/internal/input"
	"github.com/felipeelias/claude-statusline/internal/render"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		_, _ = os.Stdout.WriteString(version + "\n")
		os.Exit(0)
	}

	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}

	data, err := input.Parse(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "input error:", err)
		os.Exit(1)
	}

	output, err := render.Render(cfg, data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "render error:", err)
		os.Exit(1)
	}

	_, _ = os.Stdout.WriteString(output)
}
