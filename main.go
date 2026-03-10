package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	fmt.Fprintln(os.Stderr, "claude-statusline", version)
}
