package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// performs comprehensive validation on configuration settings
func ValidateSettings(settings *Settings) error {
	// Run struct validation
	if err := validate.Struct(settings); err != nil {
		return formatValidationError(err)
	}

	// Custom validation rules
	if err := validateMutualExclusions(settings); err != nil {
		return err
	}

	if err := validateTimingRanges(settings); err != nil {
		return err
	}

	if err := validatePresetOverrides(settings); err != nil {
		return err
	}

	return nil
}

// checks for conflicting settings
func validateMutualExclusions(s *Settings) error {
	// Save defaults
	if len(s.Save.SkipWorkspaces) > 0 && len(s.Save.OnlyWorkspaces) > 0 {
		return fmt.Errorf("cannot specify both 'save.skip_workspaces' and 'save.only_workspaces'")
	}

	if len(s.Save.SkipApps) > 0 && len(s.Save.OnlyApps) > 0 {
		return fmt.Errorf("cannot specify both 'save.skip_apps' and 'save.only_apps'")
	}

	// Load defaults
	if s.Load.Clear && s.Load.ClearAll {
		return fmt.Errorf("cannot specify both 'load.clear' and 'load.clear_all'")
	}

	if s.Load.ReuseAll && len(s.Load.ReuseApps) > 0 {
		return fmt.Errorf("'load.reuse_all' makes 'load.reuse_apps' redundant")
	}

	return nil
}

// ensures timing values are reasonable
func validateTimingRanges(s *Settings) error {
	timings := map[string]int{
		"timing.window_spawn_delay": s.Timing.WindowSpawnDelay,
		"timing.workspace_switch_delay": s.Timing.WorkspaceSwitchDelay,
		"timing.focus_settle_delay": s.Timing.FocusSettleDelay,
		"timing.post_restore_delay": s.Timing.PostRestoreDelay,
	}

	for name, value := range timings {
		if value < 0 {
			return fmt.Errorf("%s cannot be negative (got %d)", name, value)
		}
		if value > 5000 {
			return fmt.Errorf("%s is too large (got %d, max 5000ms)", name, value)
		}
	}

	return nil
}

// validates preset-specific overrides
func validatePresetOverrides(s *Settings) error {
	for presetName, override := range s.Presets {
		if presetName == "" {
			return fmt.Errorf("preset name cannot be empty")
		}

		// Validate save overrides
		if override.Save != nil {
			if len(override.Save.SkipWorkspaces) > 0 && len(override.Save.OnlyWorkspaces) > 0 {
				return fmt.Errorf("preset '%s': cannot specify both skip_workspaces and only_workspaces", presetName)
			}
			if len(override.Save.SkipApps) > 0 && len(override.Save.OnlyApps) > 0 {
				return fmt.Errorf("preset '%s': cannot specify both skip_apps and only_apps", presetName)
			}
		}

		// Validate load overrides
		if override.Load != nil {
			if override.Load.Clear && override.Load.ClearAll {
				return fmt.Errorf("preset '%s': cannot specify both clear and clear_all", presetName)
			}
			if override.Load.ReuseAll && len(override.Load.ReuseApps) > 0 {
				return fmt.Errorf("preset '%s': reuse_all makes reuse_apps redundant", presetName)
			}
		}
	}

	return nil
}

// formats validator errors in a user-friendly way
func formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			switch e.Tag() {
			case "required":
				messages = append(messages, fmt.Sprintf("field '%s' is required", field))
			case "min":
				messages = append(messages, fmt.Sprintf("field '%s' must be at least %s", field, e.Param()))
			case "max":
				messages = append(messages, fmt.Sprintf("field '%s' must be at most %s", field, e.Param()))
			default:
				messages = append(messages, fmt.Sprintf("field '%s' failed validation: %s", field, e.Tag()))
			}
		}
		return fmt.Errorf("validation errors: %s", strings.Join(messages, "; "))
	}
	return err
}