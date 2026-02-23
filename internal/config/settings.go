package config

import _ "embed"

//go:embed config.yaml
var defaultConfigTemplate string

// this is the complete configuration structure
type Settings struct {
	General GeneralSettings `yaml:"general"`
	Save SaveDefaults `yaml:"save"`
	Load LoadDefaults `yaml:"load"`
	Timing TimingSettings `yaml:"timing"`
	Presets map[string]PresetOverride `yaml:"presets"`
}

type GeneralSettings struct {
	DataDir string `yaml:"data_dir"`
	DefaultPresetName string `yaml:"default_preset_name" validate:"required"`
}

type SaveDefaults struct {
	Overwrite bool `yaml:"overwrite"`
	SkipWorkspaces []string `yaml:"skip_workspaces"`
	OnlyWorkspaces []string `yaml:"only_workspaces"`
	SkipApps []string `yaml:"skip_apps"`
	OnlyApps []string `yaml:"only_apps"`
	IgnoreFloating bool `yaml:"ignore_floating"`
	ExportFormat string `yaml:"export_format"`
}

type LoadDefaults struct {
	Clear bool `yaml:"clear"`
	ClearAll bool `yaml:"clear_all"`
	ReuseAll bool `yaml:"reuse_all"`
	ReuseApps []string `yaml:"reuse_apps"`
	FocusWorkspace string `yaml:"focus_workspace"`
	FocusApp string `yaml:"focus_app"`
}

type TimingSettings struct {
	WindowSpawnDelay int `yaml:"window_spawn_delay" validate:"min=0,max=5000"`
	WorkspaceSwitchDelay int `yaml:"workspace_switch_delay" validate:"min=0,max=5000"`
	FocusSettleDelay int `yaml:"focus_settle_delay" validate:"min=0,max=5000"`
	PostRestoreDelay int `yaml:"post_restore_delay" validate:"min=0,max=5000"`
}

type PresetOverride struct {
	Save *SaveDefaults `yaml:"save,omitempty"`
	Load *LoadDefaults `yaml:"load,omitempty"`
}

func DefaultSettings() *Settings {
	return &Settings{
		General: GeneralSettings{
			DataDir: "",
			DefaultPresetName: "layout-%Y-%m-%d-%H-%M-%S",
		},
		Save: SaveDefaults{
			Overwrite: false,
			SkipWorkspaces: []string{},
			OnlyWorkspaces: []string{},
			SkipApps: []string{},
			OnlyApps: []string{},
			IgnoreFloating: false,
			ExportFormat: "",
		},
		Load: LoadDefaults{
			Clear: false,
			ClearAll: false,
			ReuseAll: false,
			ReuseApps: []string{},
			FocusWorkspace: "",
			FocusApp: "",
		},
		Timing: TimingSettings{
			WindowSpawnDelay: 400,
			WorkspaceSwitchDelay: 100,
			FocusSettleDelay: 100,
			PostRestoreDelay: 200,
		},
		Presets: make(map[string]PresetOverride),
	}
}

// returns a YAML embedded default configuration template
func DefaultConfigTemplate() string {
	return defaultConfigTemplate
}
