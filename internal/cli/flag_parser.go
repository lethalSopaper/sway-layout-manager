package cli

import (
	"fmt"
	"os"
	"strings"
)

// holds all parsed command-line flags
type CommandFlags struct {
	SkipWorkspaces []string // for --skip-workspace=
	OnlyWorkspaces []string // for --only-workspace=
	SkipApps []string // for --skip-app=
	OnlyApps []string // for --only-app=
	FocusWorkspace string // for --focus-workspace=
	FocusApp string // for --focus-app=
	Export string // for --export=
	Import string // for --import=
	Overwrite bool // for --overwrite
	IgnoreFloating bool // for --ignore-floating
	ReuseAll bool // for --reuse-all
	ReuseApps []string // for --reuse=
	Clear bool // for --clear
	ClearAll bool // for --clear-all
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
				// handle standalone flags that cause immediate exit
				if HandleStandaloneFlag(arg) {
					continue
				}
				// handle boolean flags that set state
				if parseBooleanFlag(flags, arg) {
					continue
				}
				// unknown flag
				if isKnownValueFlag(arg) {
					fmt.Fprintf(os.Stderr, "Error: Flag '%s' requires a value (use --flag=value format)\n", arg)
				} else {
					fmt.Fprintf(os.Stderr, "Error: Unknown flag '%s'\n", arg)
				}
				os.Exit(1)
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
	case "--skip-workspace", "--only-workspace", "--skip-app", "--only-app", "--focus-workspace", "--focus-app", "--export", "--import", "--reuse":
		return true
	default:
		return false
	}
}

// handles boolean flags that set state
func parseBooleanFlag(flags *CommandFlags, flagName string) bool {
	switch flagName {
	case "--overwrite":
		flags.Overwrite = true
		return true
	case "--ignore-floating":
		flags.IgnoreFloating = true
		return true
	case "--reuse-all":
		flags.ReuseAll = true
		return true
	case "--clear":
		flags.Clear = true
		return true
	case "--clear-all":
		flags.ClearAll = true
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
	case "--only-workspace":
		// split by comma and trim whitespace
		identifiers := strings.Split(flagValue, ",")
		for _, id := range identifiers {
			id = strings.TrimSpace(id)
			if id != "" {
				flags.OnlyWorkspaces = append(flags.OnlyWorkspaces, id)
			}
		}
		return true
	case "--skip-app":
		// split by comma and trim whitespace
		identifiers := strings.Split(flagValue, ",")
		for _, id := range identifiers {
			id = strings.TrimSpace(id)
			if id != "" {
				flags.SkipApps = append(flags.SkipApps, id)
			}
		}
		return true
	case "--only-app":
		// split by comma and trim whitespace
		identifiers := strings.Split(flagValue, ",")
		for _, id := range identifiers {
			id = strings.TrimSpace(id)
			if id != "" {
				flags.OnlyApps = append(flags.OnlyApps, id)
			}
		}
		return true
	case "--focus-workspace":
		flags.FocusWorkspace = strings.TrimSpace(flagValue)
		return true
	case "--focus-app":
		flags.FocusApp = strings.TrimSpace(flagValue)
		return true
	case "--export":
		flags.Export = strings.TrimSpace(flagValue)
		return true
	case "--import":
		flags.Import = strings.TrimSpace(flagValue)
		return true
	case "--reuse":
		identifiers := strings.Split(flagValue, ",")
		for _, id := range identifiers {
			id = strings.TrimSpace(id)
			if id != "" {
				flags.ReuseApps = append(flags.ReuseApps, id)
			}
		}
		return true
	default:
		return false
	}
}
