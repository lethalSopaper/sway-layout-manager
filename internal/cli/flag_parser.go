package cli

import (
	"fmt"
	"os"
	"strings"
)

// holds all parsed command-line flags
type CommandFlags struct {
	SkipWorkspaces []string // for --skip-workspace=
}

// parses command-line arguments and returns flags, command, and remaining args
func ParseFlags(args []string) (*CommandFlags, string, []string) {
	flags := &CommandFlags{}
	var command string
	var remainingArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// check for flags
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			// handle --flag=value format
			if strings.Contains(arg, "=") {
				parts := strings.SplitN(arg, "=", 2)
				flagName := parts[0]
				flagValue := parts[1]

				if !parseValueFlag(flags, flagName, flagValue) {
					fmt.Fprintf(os.Stderr, "Error: Unknown flag '%s'\n", flagName)
					os.Exit(1)
				}
			} else {
				// handle standalone flags
				handled := HandleStandaloneFlag(arg)
				if !handled {
					if isKnownValueFlag(arg) {
						fmt.Fprintf(os.Stderr, "Error: Flag '%s' requires a value (use --flag=value format)\n", arg)
					} else {
						fmt.Fprintf(os.Stderr, "Error: Unknown flag '%s'\n", arg)
					}
					os.Exit(1)
				}
			}
		} else if command == "" {
			command = arg
		} else {
			remainingArgs = append(remainingArgs, arg)
		}
	}

	return flags, command, remainingArgs
}

// checks if a flag name is a recognized flag that requires a value
func isKnownValueFlag(flagName string) bool {
	switch flagName {
	case "--skip-workspace":
		return true
	default:
		return false
	}
}

// handles flags that take values (--flag=value format)
func parseValueFlag(flags *CommandFlags, flagName string, flagValue string) bool {
	switch flagName {
	case "--skip-workspace":
		// split by comma and trim whitespace
		identifiers := strings.Split(flagValue, ",")
		for _, id := range identifiers {
			id = strings.TrimSpace(id)
			if id != "" {
				flags.SkipWorkspaces = append(flags.SkipWorkspaces, id)
			}
		}
		return true
	default:
		return false
	}
}
