package main

import (
	"fmt"
	"os"

	"github.com/h13/gtm-users/internal/cli"
)

// version is set by goreleaser via ldflags.
var version = "dev"

func main() {
	if err := cli.NewRootCmd(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
