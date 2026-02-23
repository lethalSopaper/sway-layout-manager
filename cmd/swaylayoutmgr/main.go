package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

	flags, command, args := cli.ParseFlags(os.Args[1:])

	// handle init command
	if command == "init" {
		handleInit(cfg, flags)
		return
	}

	swayClient, err := sway.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to connect to Sway: %v\n", err)
		fmt.Fprintf(os.Stderr, "Make sure you're running this under Sway window manager.\n")
		os.Exit(1)
	}
	parser := layout.NewParser(swayClient)
	manager := preset.NewManager(cfg)

	// merge config with flags
	presetName := ""
	if len(args) > 0 {
		presetName = args[0]
	}
	flags = config.MergeConfig(cfg.Settings, flags, command, presetName)

	// command parser
	switch command {
	case "save":
		handleSave(parser, manager, flags, args)
	case "load":
		handleLoad(manager, swayClient, flags, args)
	case "list":
		handleList(manager, flags)
	case "delete":
		handleDelete(manager, flags, args)
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
		if !flags.Overwrite {
			fmt.Fprintf(os.Stderr, "Error: Preset '%s' already exists\n", name)
			fmt.Fprintf(os.Stderr, "Use a different name, delete it first with: swaylayoutmgr delete %s\n", name)
			fmt.Fprintf(os.Stderr, "Or use --overwrite to replace it.\n")
			os.Exit(1)
		}
		fmt.Printf("Overwriting existing preset '%s'...\n", name)
	}

	if !manager.Exists(name) {
		fmt.Printf("Capturing current layout as '%s'...\n", name)
	}

	// capture current layout
	preset, err := parser.CaptureCurrentLayout(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to capture layout: %v\n", err)
		os.Exit(1)
	}

	// check for conflicting flags
	if len(flags.SkipWorkspaces) > 0 && len(flags.OnlyWorkspaces) > 0 {
		fmt.Fprintf(os.Stderr, "Error: Cannot use both --skip-workspace and --only-workspace flags together\n")
		os.Exit(1)
	}

	// filter workspaces based on flags
	if len(flags.SkipWorkspaces) > 0 {
		// warn about non-existent workspace identifiers
		if err := validateWorkspaceIdentifiers(preset.Workspaces, flags.SkipWorkspaces); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}

		originalCount := len(preset.Workspaces)
		preset.Workspaces = layout.FilterSkipWorkspaces(preset.Workspaces, flags.SkipWorkspaces)
		skipped := originalCount - len(preset.Workspaces)
		if skipped > 0 {
			fmt.Printf("Skipped %d specified workspace(s)\n", skipped)
		}
	} else if len(flags.OnlyWorkspaces) > 0 {
		// check which workspace identifiers exist
		validWorkspaces, invalidWorkspaces := filterValidWorkspaces(preset.Workspaces, flags.OnlyWorkspaces)

		// error if all workspaces are invalid
		if len(validWorkspaces) == 0 {
			fmt.Fprintf(os.Stderr, "Error: workspace(s) not found: %s\n", strings.Join(invalidWorkspaces, ", "))
			os.Exit(1)
		}

		// warn about invalid workspaces but continue with valid ones
		if len(invalidWorkspaces) > 0 {
			fmt.Fprintf(os.Stderr, "Warning: workspace(s) not found: %s\n", strings.Join(invalidWorkspaces, ", "))
		}

		preset.Workspaces = layout.FilterOnlyWorkspaces(preset.Workspaces, validWorkspaces)
		included := len(preset.Workspaces)
		if included > 0 {
			fmt.Printf("Including only %d specified workspace(s)\n", included)
		}
	}

	// filter floating windows if flag is set
	if flags.IgnoreFloating {
		preset.Workspaces = layout.FilterFloatingContainers(preset.Workspaces)
		fmt.Printf("Ignoring floating windows\n")
	}

	// filter applications if flags are set
	if len(flags.SkipApps) > 0 {
		originalCount := countContainers(preset.Workspaces)
		preset.Workspaces = layout.FilterSkipApps(preset.Workspaces, flags.SkipApps)
		skipped := originalCount - countContainers(preset.Workspaces)
		if skipped > 0 {
			fmt.Printf("Skipped %d specified application(s)\n", skipped)
		}
	} else if len(flags.OnlyApps) > 0 {
		preset.Workspaces = layout.FilterOnlyApps(preset.Workspaces, flags.OnlyApps)
		included := countContainers(preset.Workspaces)
		if included > 0 {
			fmt.Printf("Including only %d specified application(s)\n", included)
		}
	}

	// export to custom path or save to presets directory
	if flags.Export != "" {
		filePath, err := exportToFile(preset, flags.Export, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to export layout: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully exported layout to '%s'\n", filePath)
	} else {
		// saves the preset to default directory
		if err := manager.Save(preset); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to save preset: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully saved layout preset '%s'\n", name)
	}

	// count total containers
	totalContainers := 0
	for _, ws := range preset.Workspaces {
		totalContainers += len(ws.Containers)
	}

	fmt.Printf("- Outputs: %d\n", len(preset.Outputs))
	fmt.Printf("- Workspaces: %d\n", len(preset.Workspaces))
	fmt.Printf("- Containers: %d\n", totalContainers)

	// warn about incompatible flags
	warnIncompatibleFlags(flags, "save")
}

func handleLoad(manager *preset.Manager, swayClient *sway.Client, flags *cli.CommandFlags, args []string) {
	var preset *layout.Preset
	var err error
	var name string

	// import from custom path or load from presets directory
	if flags.Import != "" {
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "Error: Missing preset name\n")
			fmt.Fprintf(os.Stderr, "Usage: swaylayoutmgr --import=<directory> load <name>\n")
			os.Exit(1)
		}
		name = args[0]

		importedPreset, err := importFromFile(flags.Import, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to import layout: %v\n", err)
			os.Exit(1)
		}
		preset = &importedPreset
		fmt.Printf("Imported layout from '%s'\n", filepath.Join(flags.Import, name+".json"))
	} else {
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "Error: Missing preset name\n")
			fmt.Fprintf(os.Stderr, "Usage: swaylayoutmgr load <name>\n")
			fmt.Fprintf(os.Stderr, "\nRun 'swaylayoutmgr list' to see available presets.\n")
			os.Exit(1)
		}
		name = args[0]

		// check if preset exists
		if !manager.Exists(name) {
			fmt.Fprintf(os.Stderr, "Error: Preset '%s' not found\n", name)
			fmt.Fprintf(os.Stderr, "Run 'swaylayoutmgr list' to see available presets.\n")
			os.Exit(1)
		}

		// load the preset
		preset, err = manager.Load(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to load preset: %v\n", err)
			os.Exit(1)
		}
	}

	// check for conflicting flags
	if len(flags.SkipWorkspaces) > 0 && len(flags.OnlyWorkspaces) > 0 {
		fmt.Fprintf(os.Stderr, "Error: Cannot use both --skip-workspace and --only-workspace flags together\n")
		os.Exit(1)
	}

	if flags.Clear && flags.ClearAll {
		fmt.Fprintf(os.Stderr, "Error: Cannot use both --clear and --clear-all flags together\n")
		os.Exit(1)
	}

	if flags.FocusWorkspace != "" && flags.FocusApp != "" {
		fmt.Fprintf(os.Stderr, "Error: Cannot use both --focus-workspace and --focus-app flags together\n")
		os.Exit(1)
	}

	// filter workspaces based on flags
	if len(flags.SkipWorkspaces) > 0 {
		// warn about non-existent workspace identifiers
		if err := validateWorkspaceIdentifiers(preset.Workspaces, flags.SkipWorkspaces); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}

		originalCount := len(preset.Workspaces)
		preset.Workspaces = layout.FilterSkipWorkspaces(preset.Workspaces, flags.SkipWorkspaces)
		skipped := originalCount - len(preset.Workspaces)
		if skipped > 0 {
			fmt.Printf("Skipping %d specified workspace(s)\n", skipped)
		}
	} else if len(flags.OnlyWorkspaces) > 0 {
		// check which workspace identifiers exist
		validWorkspaces, invalidWorkspaces := filterValidWorkspaces(preset.Workspaces, flags.OnlyWorkspaces)

		// error if all workspaces are invalid
		if len(validWorkspaces) == 0 {
			fmt.Fprintf(os.Stderr, "Error: workspace(s) not found: %s\n", strings.Join(invalidWorkspaces, ", "))
			os.Exit(1)
		}

		// warn about invalid workspaces but continue with valid ones
		if len(invalidWorkspaces) > 0 {
			fmt.Fprintf(os.Stderr, "Warning: workspace(s) not found: %s\n", strings.Join(invalidWorkspaces, ", "))
		}

		preset.Workspaces = layout.FilterOnlyWorkspaces(preset.Workspaces, validWorkspaces)
		included := len(preset.Workspaces)
		if included > 0 {
			fmt.Printf("Restoring only %d specified workspace(s)\n", included)
		}
	}

	// filter floating windows if flag is set
	if flags.IgnoreFloating {
		preset.Workspaces = layout.FilterFloatingContainers(preset.Workspaces)
		fmt.Printf("Ignoring floating windows\n")
	}
	// filter applications if flags are set
	if len(flags.SkipApps) > 0 {
		originalCount := countContainers(preset.Workspaces)
		preset.Workspaces = layout.FilterSkipApps(preset.Workspaces, flags.SkipApps)
		skipped := originalCount - countContainers(preset.Workspaces)
		if skipped > 0 {
			fmt.Printf("Skipping %d specified application(s)\n", skipped)
		}
	} else if len(flags.OnlyApps) > 0 {
		preset.Workspaces = layout.FilterOnlyApps(preset.Workspaces, flags.OnlyApps)
		included := countContainers(preset.Workspaces)
		if included > 0 {
			fmt.Printf("Restoring only %d specified application(s)\n", included)
		}
	}

	// warn if both reuse flags are used together
	if flags.ReuseAll && len(flags.ReuseApps) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: --reuse-all is set, ignoring --reuse flag\n")
	}

	fmt.Printf("Restoring layout '%s'...\n", name)

	// restore the layout
	restorer := layout.NewRestorer(swayClient)
	restorer.ReuseExistingWindows = flags.ReuseAll

	if !flags.ReuseAll {
		restorer.ReuseApps = flags.ReuseApps
	}
	restorer.ClearWorkspaces = flags.Clear
	restorer.ClearAllWorkspaces = flags.ClearAll
	result, err := restorer.Restore(preset)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to restore layout: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nLayout '%s' restoration complete:\n", name)
	fmt.Printf("- Workspaces restored: %d\n", result.WorkspacesRestored)
	fmt.Printf("- Applications launched: %d\n", result.ApplicationsLaunched)
	if result.ApplicationsReused > 0 {
		fmt.Printf("- Applications reused: %d\n", result.ApplicationsReused)
	}

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

	// focus workspace if flag is set
	if flags.FocusWorkspace != "" {
		if err := focusWorkspace(swayClient, flags.FocusWorkspace); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to focus workspace '%s': %v\n", flags.FocusWorkspace, err)
		} else {
			fmt.Printf("\nFocused workspace '%s'\n", flags.FocusWorkspace)
		}
	}

	// focus application if flag is set
	if flags.FocusApp != "" {
		if err := focusApplication(swayClient, flags.FocusApp); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to focus application '%s': %v\n", flags.FocusApp, err)
		} else {
			fmt.Printf("\nFocused application '%s'\n", flags.FocusApp)
		}
	}

	// warn about incompatible flags
	warnIncompatibleFlags(flags, "load")
}

func handleList(manager *preset.Manager, flags *cli.CommandFlags) {
	// warn about incompatible flags
	warnIncompatibleFlags(flags, "list")

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

func handleDelete(manager *preset.Manager, flags *cli.CommandFlags, args []string) {
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

	// warn about incompatible flags
	warnIncompatibleFlags(flags, "delete")
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

// separates workspace identifiers into valid and invalid lists
func filterValidWorkspaces(workspaces []layout.WorkspaceLayout, identifiers []string) (valid []string, invalid []string) {
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

		if found {
			valid = append(valid, identifier)
		} else {
			invalid = append(invalid, identifier)
		}
	}

	return valid, invalid
}

// counts total number of containers across all workspaces
func countContainers(workspaces []layout.WorkspaceLayout) int {
	count := 0
	for _, ws := range workspaces {
		count += countContainersRecursive(ws.Containers)
	}
	return count
}

func countContainersRecursive(containers []layout.ContainerLayout) int {
	count := len(containers)
	for _, container := range containers {
		if len(container.Children) > 0 {
			count += countContainersRecursive(container.Children)
		}
	}
	return count
}

// focuses the specified workspace by number or name
func focusWorkspace(swayClient *sway.Client, identifier string) error {
	// check if identifier is a workspace number
	if _, err := strconv.Atoi(identifier); err == nil {
		return swayClient.RunCommand(fmt.Sprintf("workspace number %s", identifier))
	}
	return swayClient.RunCommand(fmt.Sprintf("workspace %s", identifier))
}

// focuses the specified application by app_id or class
func focusApplication(swayClient *sway.Client, identifier string) error {
	tree, err := swayClient.GetTree()
	if err != nil {
		return fmt.Errorf("failed to get window tree: %w", err)
	}

	// search for the application in the tree
	if !findApplicationInTree(tree, identifier) {
		return fmt.Errorf("application '%s' not found", identifier)
	}

	// try focusing by app_id first
	err = swayClient.RunCommand(fmt.Sprintf("[app_id=\"%s\"] focus", identifier))
	if err == nil {
		return nil
	}

	err = swayClient.RunCommand(fmt.Sprintf("[class=\"%s\"] focus", identifier))
	return err
}

// recursively searches for an application in the window tree
func findApplicationInTree(node *sway.Node, identifier string) bool {
	// check if this node matches by app_id
	if node.AppID == identifier {
		return true
	}

	// check if this node matches by window class (X11)
	if node.WindowProperties != nil && node.WindowProperties.Class == identifier {
		return true
	}

	// search children
	for i := range node.Nodes {
		if findApplicationInTree(&node.Nodes[i], identifier) {
			return true
		}
	}
	for i := range node.FloatingNodes {
		if findApplicationInTree(&node.FloatingNodes[i], identifier) {
			return true
		}
	}

	return false
}

// exports a preset to a custom file path
func exportToFile(preset *layout.Preset, dirPath string, name string) (string, error) {
	// validate that the directory exists or create it
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("directory does not exist: %s", dirPath)
		}
		return "", fmt.Errorf("failed to access directory %s: %w", dirPath, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", dirPath)
	}

	// construct full file path
	filePath := filepath.Join(dirPath, name+".json")

	data, err := json.MarshalIndent(preset, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal preset to JSON: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return filePath, nil
}

// imports a preset from a custom directory path
func importFromFile(dirPath string, name string) (layout.Preset, error) {
	// validate that the directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return layout.Preset{}, fmt.Errorf("failed to access directory %s: %w", dirPath, err)
	}
	if !info.IsDir() {
		return layout.Preset{}, fmt.Errorf("path is not a directory: %s", dirPath)
	}

	// construct full file path
	filePath := filepath.Join(dirPath, name+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return layout.Preset{}, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var preset layout.Preset
	if err := json.Unmarshal(data, &preset); err != nil {
		return layout.Preset{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return preset, nil
}

// warns about flags that don't apply to the given command
func warnIncompatibleFlags(flags *cli.CommandFlags, command string) {
	flagCompatibility := map[string][]string{
		"--skip-workspace": {"save", "load"},
		"--only-workspace": {"save", "load"},
		"--skip-app": {"save", "load"},
		"--only-app": {"save", "load"},
		"--focus-workspace": {"load"},
		"--focus-app": {"load"},
		"--overwrite": {"save"},
		"--ignore-floating": {"save", "load"},
		"--export": {"save"},
		"--import": {"load"},
		"--reuse-all": {"load"},
		"--reuse": {"load"},
		"--clear": {"load"},
		"--clear-all": {"load"},
	}

	type flagCheck struct {
		name   string
		active bool
	}

	flagChecks := []flagCheck{
		{"--skip-workspace", len(flags.SkipWorkspaces) > 0},
		{"--only-workspace", len(flags.OnlyWorkspaces) > 0},
		{"--skip-app", len(flags.SkipApps) > 0},
		{"--only-app", len(flags.OnlyApps) > 0},
		{"--focus-workspace", flags.FocusWorkspace != ""},
		{"--focus-app", flags.FocusApp != ""},
		{"--overwrite", flags.Overwrite},
		{"--ignore-floating", flags.IgnoreFloating},
		{"--export", flags.Export != ""},
		{"--import", flags.Import != ""},
		{"--clear", flags.Clear},
		{"--clear-all", flags.ClearAll},
		{"--reuse-all", flags.ReuseAll},
		{"--reuse", len(flags.ReuseApps) > 0},
	}

	var warnings []string

	for _, check := range flagChecks {
		if check.active {
			compatibleCmds := flagCompatibility[check.name]
			if !contains(compatibleCmds, command) {
				warnings = append(warnings, formatFlagWarning(check.name, compatibleCmds))
			}
		}
	}

	if len(warnings) > 0 {
		fmt.Fprintf(os.Stderr, "\n")
		for _, warning := range warnings {
			fmt.Fprintf(os.Stderr, "Warning: %s (ignored)\n", warning)
		}
	}
}

// checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generates a default configuration file
func handleInit(cfg *config.Config, flags *cli.CommandFlags) {
	configPath := cfg.ConfigPath()

	// check if config already exists
	if _, err := os.Stat(configPath); err == nil && !flags.Overwrite {
		fmt.Fprintf(os.Stderr, "Error: Config file already exists at %s\n", configPath)
		fmt.Fprintf(os.Stderr, "Use --overwrite to replace it.\n")
		os.Exit(1)
	}

	// generate default config
	if err := config.GenerateDefaultConfig(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created default configuration at %s\n", configPath)
	fmt.Printf("Edit this file to customize your defaults.\n")
}

// formats a warning message for incompatible flags
func formatFlagWarning(flagName string, compatibleCommands []string) string {
	if len(compatibleCommands) == 1 {
		return fmt.Sprintf("%s only works with '%s' command", flagName, compatibleCommands[0])
	}
	cmdList := strings.Join(compatibleCommands, "' and '")
	return fmt.Sprintf("%s only works with '%s' commands", flagName, cmdList)
}

func printUsage() {
	fmt.Print(helpText)
}
