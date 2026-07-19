package main

import (
	"fmt"
	"os"

	"github.com/florianmueller/ordr/internal/cli"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	if err := cli.Execute(version); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
