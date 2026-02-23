package config

import (
	"github.com/lidsol/sway-layout-manager/internal/cli"
)

func MergeConfig(settings *Settings, cliFlags *cli.CommandFlags, command string, presetName string) *cli.CommandFlags {
	// Start with CLI flags as base (they have highest priority)
	merged := &cli.CommandFlags{
		SkipWorkspaces: cliFlags.SkipWorkspaces,
		OnlyWorkspaces: cliFlags.OnlyWorkspaces,
		SkipApps: cliFlags.SkipApps,
		OnlyApps: cliFlags.OnlyApps,
		FocusWorkspace: cliFlags.FocusWorkspace,
		FocusApp: cliFlags.FocusApp,
		Export: cliFlags.Export,
		Import: cliFlags.Import,
		Overwrite: cliFlags.Overwrite,
		IgnoreFloating: cliFlags.IgnoreFloating,
		ReuseAll: cliFlags.ReuseAll,
		ReuseApps: cliFlags.ReuseApps,
		Clear: cliFlags.Clear,
		ClearAll: cliFlags.ClearAll,
	}

	// Get preset-specific overrides if available
	var presetOverride *PresetOverride
	if presetName != "" {
		if override, exists := settings.Presets[presetName]; exists {
			presetOverride = &override
		}
	}

	// Apply defaults based on command type
	switch command {
	case "save":
		merged = applySaveDefaults(merged, settings, presetOverride, cliFlags)
	case "load":
		merged = applyLoadDefaults(merged, settings, presetOverride, cliFlags)
	}

	return merged
}

// applies save-specific configuration defaults
func applySaveDefaults(merged *cli.CommandFlags, settings *Settings, presetOverride *PresetOverride, cliFlags *cli.CommandFlags) *cli.CommandFlags {
	saveDefaults := &settings.Save
	if presetOverride != nil && presetOverride.Save != nil {
		saveDefaults = presetOverride.Save
	}

	// Apply defaults only if CLI flags were not explicitly set
	if !cliFlags.Overwrite && len(cliFlags.SkipWorkspaces) == 0 && len(cliFlags.OnlyWorkspaces) == 0 &&
		len(cliFlags.SkipApps) == 0 && len(cliFlags.OnlyApps) == 0 &&
		!cliFlags.IgnoreFloating && cliFlags.Export == "" {

		// Only apply config defaults if no related CLI flags were provided
		if len(cliFlags.SkipWorkspaces) == 0 && len(cliFlags.OnlyWorkspaces) == 0 {
			merged.SkipWorkspaces = saveDefaults.SkipWorkspaces
			merged.OnlyWorkspaces = saveDefaults.OnlyWorkspaces
		}

		if len(cliFlags.SkipApps) == 0 && len(cliFlags.OnlyApps) == 0 {
			merged.SkipApps = saveDefaults.SkipApps
			merged.OnlyApps = saveDefaults.OnlyApps
		}

		if !cliFlags.IgnoreFloating {
			merged.IgnoreFloating = saveDefaults.IgnoreFloating
		}

		if cliFlags.Export == "" {
			merged.Export = saveDefaults.ExportFormat
		}

		// only apply if not explicitly set via CLI
		if !wasOverwriteFlagProvided(cliFlags) {
			merged.Overwrite = saveDefaults.Overwrite
		}
	}

	return merged
}

// applies load-specific configuration defaults
func applyLoadDefaults(merged *cli.CommandFlags, settings *Settings, presetOverride *PresetOverride, cliFlags *cli.CommandFlags) *cli.CommandFlags {
	loadDefaults := &settings.Load
	if presetOverride != nil && presetOverride.Load != nil {
		loadDefaults = presetOverride.Load
	}

	// Apply defaults only if CLI flags were not explicitly set
	if !cliFlags.Clear && !wasClearFlagProvided(cliFlags) {
		merged.Clear = loadDefaults.Clear
	}

	if !cliFlags.ClearAll && !wasClearAllFlagProvided(cliFlags) {
		merged.ClearAll = loadDefaults.ClearAll
	}

	if !cliFlags.ReuseAll && !wasReuseAllFlagProvided(cliFlags) {
		merged.ReuseAll = loadDefaults.ReuseAll
	}

	if len(cliFlags.ReuseApps) == 0 {
		merged.ReuseApps = loadDefaults.ReuseApps
	}

	if cliFlags.FocusWorkspace == "" {
		merged.FocusWorkspace = loadDefaults.FocusWorkspace
	}

	if cliFlags.FocusApp == "" {
		merged.FocusApp = loadDefaults.FocusApp
	}

	return merged
}

func wasOverwriteFlagProvided(flags *cli.CommandFlags) bool {
	return flags.Overwrite
}

func wasClearFlagProvided(flags *cli.CommandFlags) bool {
	return flags.Clear
}

func wasClearAllFlagProvided(flags *cli.CommandFlags) bool {
	return flags.ClearAll
}

func wasReuseAllFlagProvided(flags *cli.CommandFlags) bool {
	return flags.ReuseAll
}

func GetTimingSettings(settings *Settings) TimingSettings {
	return settings.Timing
}
