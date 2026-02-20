package cli

import (
	"fmt"
	"os"
)

// version will be set by main package
var AppVersion string
var HelpText string

// HandleStandaloneFlag processes standalone boolean flags (no value required)
// These flags control program behavior directly without data processing
// Examples: --help, --version, --overwrite, --verbose
// Returns true if the flag was recognized and handled
func HandleStandaloneFlag(flag string) bool {
	switch flag {
	case "--help", "-h":
		printHelp()
		os.Exit(0)
		return true
	case "--version", "-v":
		printVersion()
		os.Exit(0)
		return true
	default:
		return false
	}
}

func printHelp() {
	fmt.Print(HelpText)
}

func printVersion() {
	fmt.Printf("sway-layout-manager %s\n", AppVersion)
}
