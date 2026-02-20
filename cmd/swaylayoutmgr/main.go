package main

import (
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"github.com/lidsol/sway-layout-manager/internal/cli"
	"github.com/lidsol/sway-layout-manager/internal/config"
	"github.com/lidsol/sway-layout-manager/internal/layout"
	"github.com/lidsol/sway-layout-manager/internal/sway"
	"github.com/lidsol/sway-layout-manager/pkg/preset"
)

const version = "0.1.0-dev"

//go:embed help.txt
var helpText string

func main() {
	cli.AppVersion = version
	cli.HelpText = helpText
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: No command specified\n\n")
		printUsage()
		os.Exit(1)
	}

	// initialize components
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	swayClient, err := sway.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to connect to Sway: %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure you're running this under Sway window manager.\n")
		os.Exit(1)
	}
	parser := layout.NewParser(swayClient)
	manager := preset.NewManager(cfg)

	// parse flags and command
	flags, command, args := cli.ParseFlags(os.Args[1:])

	// command parser
	switch command {
	case "save":
		handleSave(parser, manager, flags, args)
	case "load":
		handleLoad(manager, swayClient, flags, args)
	case "list":
		handleList(manager)
	case "delete":
		handleDelete(manager, args)
	case "":
		// no command provided
		fmt.Fprintf(os.Stderr, "Error: No command specified\n\n")
		printUsage()
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n", command)
		fmt.Fprintf(os.Stderr, "Run 'swaylayoutmgr --help' to see available commands.\n")
		os.Exit(1)
	}
}

func handleSave(parser *layout.Parser, manager *preset.Manager, flags *cli.CommandFlags, args []string) {
	var name string
	if len(args) > 0 {
		name = args[0]
	} else {
		// generates a default name
		name = fmt.Sprintf("layout-%s", time.Now().Format("2006-01-02-15-04-05"))
	}

	// check if preset already exists
	if manager.Exists(name) {
		fmt.Fprintf(os.Stderr, "Error: Preset '%s' already exists\n", name)
		fmt.Fprintf(os.Stderr, "Use a different name or delete it first with: swaylayoutmgr delete %s\n", name)
		os.Exit(1)
	}

	fmt.Printf("Capturing current layout as '%s'...\n", name)

	// capture current layout
	preset, err := parser.CaptureCurrentLayout(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to capture layout: %v\n", err)
		os.Exit(1)
	}

	// filter specific workspaces if flag is set
	if len(flags.SkipWorkspaces) > 0 {
		// validate workspace identifiers exist
		if err := validateWorkspaceIdentifiers(preset.Workspaces, flags.SkipWorkspaces); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		originalCount := len(preset.Workspaces)
		preset.Workspaces = layout.FilterSkipWorkspaces(preset.Workspaces, flags.SkipWorkspaces)
		skipped := originalCount - len(preset.Workspaces)
		if skipped > 0 {
			fmt.Printf("Skipped %d specified workspace(s)\n", skipped)
		}
	}

	// saves the preset
	if err := manager.Save(preset); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to save preset: %v\n", err)
		os.Exit(1)
	}

	// count total containers
	totalContainers := 0
	for _, ws := range preset.Workspaces {
		totalContainers += len(ws.Containers)
	}

	fmt.Printf("Successfully saved layout preset '%s'\n", name)
	fmt.Printf("- Outputs: %d\n", len(preset.Outputs))
	fmt.Printf("- Workspaces: %d\n", len(preset.Workspaces))
	fmt.Printf("- Containers: %d\n", totalContainers)
}

func handleLoad(manager *preset.Manager, swayClient *sway.Client, flags *cli.CommandFlags, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Missing preset name\n")
		fmt.Fprintf(os.Stderr, "Usage: swaylayoutmgr load <name>\n")
		fmt.Fprintf(os.Stderr, "\nRun 'swaylayoutmgr list' to see available presets.\n")
		os.Exit(1)
	}
	name := args[0]

	// check if preset exists
	if !manager.Exists(name) {
		fmt.Fprintf(os.Stderr, "Error: Preset '%s' not found\n", name)
		fmt.Fprintf(os.Stderr, "Run 'swaylayoutmgr list' to see available presets.\n")
		os.Exit(1)
	}

	// load the preset
	preset, err := manager.Load(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load preset: %v\n", err)
		os.Exit(1)
	}

	// filter specific workspaces if flag is set
	if len(flags.SkipWorkspaces) > 0 {
		// validate workspace identifiers exist
		if err := validateWorkspaceIdentifiers(preset.Workspaces, flags.SkipWorkspaces); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		originalCount := len(preset.Workspaces)
		preset.Workspaces = layout.FilterSkipWorkspaces(preset.Workspaces, flags.SkipWorkspaces)
		skipped := originalCount - len(preset.Workspaces)
		if skipped > 0 {
			fmt.Printf("Skipping %d specified workspace(s)\n", skipped)
		}
	}

	fmt.Printf("Restoring layout '%s'...\n", name)

	// restore the layout
	restorer := layout.NewRestorer(swayClient)
	result, err := restorer.Restore(preset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to restore layout: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nLayout '%s' restoration complete:\n", name)
	fmt.Printf("- Workspaces restored: %d\n", result.WorkspacesRestored)
	fmt.Printf("- Applications launched: %d\n", result.ApplicationsLaunched)

	// failed applications
	if result.ApplicationsFailed > 0 {
		fmt.Printf("- Applications failed: %d\n", result.ApplicationsFailed)
	}

	// warnings/errors
	if len(result.Errors) > 0 {
		fmt.Printf("\nWarnings/Errors encountered:\n")
		for _, err := range result.Errors {
			fmt.Printf("  - %v\n", err)
		}
	}
}

func handleList(manager *preset.Manager) {
	metadata, err := manager.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to list presets: %v\n", err)
		os.Exit(1)
	}

	if len(metadata) == 0 {
		fmt.Println("No saved layout presets found.")
		fmt.Println("Use 'swaylayoutmgr save [name]' to save your current layout.")
		return
	}

	fmt.Printf("Available layout presets:\n\n")
	for _, meta := range metadata {
		fmt.Printf("Name: %s\n", meta.Name)
		fmt.Printf("  Created: %s\n", meta.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Sway Version: %s\n", meta.SwayVersion)
		fmt.Printf("  Workspaces: %d\n", meta.WorkspaceCount)
		fmt.Printf("  Windows: %d\n", meta.WindowCount)
		fmt.Println()
	}
}

func handleDelete(manager *preset.Manager, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Missing preset name\n")
		fmt.Fprintf(os.Stderr, "Usage: swaylayoutmgr delete <name>\n")
		fmt.Fprintf(os.Stderr, "\nRun 'swaylayoutmgr list' to see available presets.\n")
		os.Exit(1)
	}
	name := args[0]

	// check if preset exists
	if !manager.Exists(name) {
		fmt.Fprintf(os.Stderr, "Error: Preset '%s' not found\n", name)
		fmt.Fprintf(os.Stderr, "Run 'swaylayoutmgr list' to see available presets.\n")
		os.Exit(1)
	}

	// deletes the preset
	if err := manager.Delete(name); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to delete preset: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully deleted preset '%s'\n", name)
}

// validateWorkspaceIdentifiers checks if all workspace identifiers exist in the workspace list
func validateWorkspaceIdentifiers(workspaces []layout.WorkspaceLayout, identifiers []string) error {
	var notFound []string

	for _, identifier := range identifiers {
		found := false

		// check if identifier is a workspace number
		if num, err := strconv.Atoi(identifier); err == nil {
			for _, ws := range workspaces {
				if ws.Num == num {
					found = true
					break
				}
			}
		} else {
			// check if identifier is a workspace name
			for _, ws := range workspaces {
				if ws.Name == identifier {
					found = true
					break
				}
			}
		}

		if !found {
			notFound = append(notFound, identifier)
		}
	}

	if len(notFound) > 0 {
		return fmt.Errorf("workspace(s) not found: %s", strings.Join(notFound, ", "))
	}

	return nil
}

func printUsage() {
	fmt.Print(helpText)
}
