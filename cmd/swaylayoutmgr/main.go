package main

import (
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/lidsol/sway-layout-manager/internal/config"
	"github.com/lidsol/sway-layout-manager/internal/layout"
	"github.com/lidsol/sway-layout-manager/internal/sway"
	"github.com/lidsol/sway-layout-manager/pkg/preset"
)

const version = "0.1.0-dev"

//go:embed help.txt
var helpText string

func main() {
	if len(os.Args) < 2 {
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

	// command parser
	command := os.Args[1]
	switch command {
	case "save":
		handleSave(parser, manager, os.Args[2:])
	case "list":
		handleList(manager)
	case "delete":
		handleDelete(manager, os.Args[2:])
	case "--version", "-v":
		fmt.Printf("sway-layout-manager %s\n", version)
	case "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleSave(parser *layout.Parser, manager *preset.Manager, args []string) {
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
		fmt.Fprintf(os.Stderr, "Use a different name or delete the existing preset first.\n")
		os.Exit(1)
	}

	fmt.Printf("Capturing current layout as '%s'...\n", name)

	// capture current layout
	preset, err := parser.CaptureCurrentLayout(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to capture layout: %v\n", err)
		os.Exit(1)
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

func handleList(manager *preset.Manager) {
	metadata, err := manager.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to list presets: %v\n", err)
		os.Exit(1)
	}

	if len(metadata) == 0 {
		fmt.Println("No saved layout presets found.")
		fmt.Println("Use 'sway-layout-manager save [name]' to save your current layout.")
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
		fmt.Fprintf(os.Stderr, "Error: Please specify a preset name to delete\n")
		fmt.Fprintf(os.Stderr, "Usage: sway-layout-manager delete <name>\n")
		os.Exit(1)
	}
	name := args[0]

	// check if preset exists
	if !manager.Exists(name) {
		fmt.Fprintf(os.Stderr, "Error: Preset '%s' not found\n", name)
		os.Exit(1)
	}

	// deletes the preset
	if err := manager.Delete(name); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to delete preset: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully deleted preset '%s'\n", name)
}

func printUsage() {
	fmt.Print(helpText)
}