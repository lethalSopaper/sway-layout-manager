package preset

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"github.com/lidsol/sway-layout-manager/internal/config"
	"github.com/lidsol/sway-layout-manager/internal/layout"
)

// Manager handles preset storage operations
type Manager struct {
	config *config.Config
}

// constructor
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config: cfg,
	}
}

// saves a layout preset to disk
func (m *Manager) Save(preset *layout.Preset) error {
	if preset.Name == "" {
		return fmt.Errorf("preset name cannot be empty")
	}

	// generate file path
	filePath := m.config.PresetPath(preset.Name)

	// create JSON data
	data, err := json.MarshalIndent(preset, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal preset to JSON: %w", err)
	}

	// write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write preset file %s: %w", filePath, err)
	}

	return nil
}

// loads a layout preset from disk
func (m *Manager) Load(name string) (*layout.Preset, error) {
	filePath := m.config.PresetPath(name)

	// check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("preset '%s' not found", name)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read preset file %s: %w", filePath, err)
	}

	var preset layout.Preset
	if err := json.Unmarshal(data, &preset); err != nil {
		return nil, fmt.Errorf("failed to parse preset JSON: %w", err)
	}

	return &preset, nil
}

// returns metadata for all available presets
func (m *Manager) List() ([]layout.Metadata, error) {
	var metadata []layout.Metadata

	// read directory contents
	entries, err := os.ReadDir(m.config.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read presets directory: %w", err)
	}

	// process each .json file
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}

		// remove .json extension from name
		name := strings.TrimSuffix(entry.Name(), ".json")

		meta := layout.Metadata{
			Name:      name,
			CreatedAt: fileInfo.ModTime(),
		}

		// extra details from preset
		if preset, err := m.Load(name); err == nil {
			meta.SwayVersion = preset.SwayVersion
			meta.WorkspaceCount = len(preset.Workspaces)
			// count total windows
			windowCount := 0
			for _, ws := range preset.Workspaces {
				windowCount += len(ws.Containers)
			}
			meta.WindowCount = windowCount
		}

		metadata = append(metadata, meta)
	}

	return metadata, nil
}

// removes a preset file from disk
func (m *Manager) Delete(name string) error {
	filePath := m.config.PresetPath(name)

	// check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("preset '%s' not found", name)
	}

	// delete file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete preset file %s: %w", filePath, err)
	}

	return nil
}

// checks if a preset with the given name exists
func (m *Manager) Exists(name string) bool {
	filePath := m.config.PresetPath(name)
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
