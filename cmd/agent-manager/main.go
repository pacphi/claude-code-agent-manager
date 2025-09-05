package main

import (
	"fmt"
	"os"

	"github.com/pacphi/claude-code-agent-manager/internal/cli/commands"
)

var version = "dev"

func main() {
	// Create command registry with all commands
	registry := commands.NewCommandRegistry()

	// Create root command with all subcommands
	rootCmd := registry.CreateRootCommand(version)

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
