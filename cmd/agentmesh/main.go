package main

import (
	"fmt"
	"os"

	"github.com/alanjm0532/agentmesh/internal/cli/command"
)

func main() {
	if err := command.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
