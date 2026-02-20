package unit

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
	"github.com/lidsol/sway-layout-manager/internal/config"
	"github.com/lidsol/sway-layout-manager/internal/layout"
	"github.com/lidsol/sway-layout-manager/pkg/preset"
)

func createTestConfig(t *testing.T) (*config.Config, func()) {
	// create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "sway-layout-manager-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cfg := &config.Config{
		DataDir:     tempDir,
		ConfigDir:   tempDir,
		DefaultName: "layout-%s",
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return cfg, cleanup
}

func createTestPreset() *layout.Preset {
	return &layout.Preset{
		Name: "test-layout",
		CreatedAt: time.Now(),
		SwayVersion: "1.11",
		Outputs: []layout.OutputLayout{
			{
				Name: "DP-1",
				Active: true,
				Rect: layout.Rect{X: 0, Y: 0, Width: 1920, Height: 1080},
			},
		},
		Workspaces: []layout.WorkspaceLayout{
			{
				Num: 1,
				Name: "1",
				Output: "DP-1",
				Layout: "splith",
				Containers: []layout.ContainerLayout{
					{
						ID: 123,
						Type: "con",
						Layout: "none",
						PID: 123,
						ExecCommand: "test-app",
						WindowTitle: "Test Window",
						Rect: layout.Rect{X: 0, Y: 0, Width: 960, Height: 1080},
						WindowRect: layout.Rect{X: 0, Y: 0, Width: 960, Height: 1080},
						Children: []layout.ContainerLayout{},
					},
				},
			},
		},
	}
}

func TestNewManager(t *testing.T) {
	cfg, cleanup := createTestConfig(t)
	defer cleanup()

	manager := preset.NewManager(cfg)

	if manager == nil {
		t.Error("NewManager() should return a valid manager")
	}
}

func TestSavePreset(t *testing.T) {
	cfg, cleanup := createTestConfig(t)
	defer cleanup()

	manager := preset.NewManager(cfg)
	testPreset := createTestPreset()

	t.Run("successful save", func(t *testing.T) {
		err := manager.Save(testPreset)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		// verifies that file was created
		filePath := cfg.PresetPath(testPreset.Name)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Preset file was not created: %s", filePath)
		}

		// check contents
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read preset file: %v", err)
		}

		var savedPreset layout.Preset
		if err := json.Unmarshal(data, &savedPreset); err != nil {
			t.Fatalf("Failed to unmarshal preset: %v", err)
		}

		if savedPreset.Name != testPreset.Name {
			t.Errorf("Expected preset name %s, got %s", testPreset.Name, savedPreset.Name)
		}
	})

	t.Run("save preset with empty name", func(t *testing.T) {
		emptyNamePreset := createTestPreset()
		emptyNamePreset.Name = ""

		err := manager.Save(emptyNamePreset)
		if err == nil {
			t.Error("Save() should fail when preset name is empty")
		}
		if !strings.Contains(err.Error(), "name cannot be empty") {
			t.Errorf("Expected error message about empty name, got: %v", err)
		}
	})
}

func TestLoadPreset(t *testing.T) {
	cfg, cleanup := createTestConfig(t)
	defer cleanup()

	manager := preset.NewManager(cfg)
	testPreset := createTestPreset()

	t.Run("load existing preset", func(t *testing.T) {
		// save the preset first
		err := manager.Save(testPreset)
		if err != nil {
			t.Fatalf("Failed to save test preset: %v", err)
		}

		// loads the preset
		loadedPreset, err := manager.Load(testPreset.Name)
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if loadedPreset.Name != testPreset.Name {
			t.Errorf("Expected preset name %s, got %s", testPreset.Name, loadedPreset.Name)
		}

		if loadedPreset.SwayVersion != testPreset.SwayVersion {
			t.Errorf("Expected sway version %s, got %s", testPreset.SwayVersion, loadedPreset.SwayVersion)
		}

		if len(loadedPreset.Outputs) != len(testPreset.Outputs) {
			t.Errorf("Expected %d outputs, got %d", len(testPreset.Outputs), len(loadedPreset.Outputs))
		}
	})

	t.Run("load non-existent preset", func(t *testing.T) {
		_, err := manager.Load("non-existent")
		if err == nil {
			t.Error("Load() should fail for non-existent preset")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})
}

func TestDeletePreset(t *testing.T) {
	cfg, cleanup := createTestConfig(t)
	defer cleanup()

	manager := preset.NewManager(cfg)
	testPreset := createTestPreset()

	t.Run("delete existing preset", func(t *testing.T) {
		// save the preset first
		err := manager.Save(testPreset)
		if err != nil {
			t.Fatalf("Failed to save test preset: %v", err)
		}

		// verify it exists
		if !manager.Exists(testPreset.Name) {
			t.Fatal("Preset should exist before deletion")
		}

		// delete it
		err = manager.Delete(testPreset.Name)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		// verify it's deleted
		if manager.Exists(testPreset.Name) {
			t.Error("Preset should not exist after deletion")
		}
	})

	t.Run("delete non-existent preset", func(t *testing.T) {
		err := manager.Delete("non-existent")
		if err == nil {
			t.Error("Delete() should fail for non-existent preset")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}
	})
}

func TestExistsPreset(t *testing.T) {
	cfg, cleanup := createTestConfig(t)
	defer cleanup()

	manager := preset.NewManager(cfg)
	testPreset := createTestPreset()

	// there is no preset yet
	if manager.Exists(testPreset.Name) {
		t.Error("Preset should not exist initially")
	}

	// save the preset
	err := manager.Save(testPreset)
	if err != nil {
		t.Fatalf("Failed to save test preset: %v", err)
	}

	// should exist now
	if !manager.Exists(testPreset.Name) {
		t.Error("Preset should exist after saving")
	}
}

func TestListPresets(t *testing.T) {
	cfg, cleanup := createTestConfig(t)
	defer cleanup()

	manager := preset.NewManager(cfg)

	t.Run("empty list", func(t *testing.T) {
		metadata, err := manager.List()
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(metadata) != 0 {
			t.Errorf("Expected empty list, got %d items", len(metadata))
		}
	})

	t.Run("list with presets", func(t *testing.T) {
		// save some presets
		preset1 := createTestPreset()
		preset1.Name = "layout-1"

		preset2 := createTestPreset()
		preset2.Name = "layout-2"

		err := manager.Save(preset1)
		if err != nil {
			t.Fatalf("Failed to save preset1: %v", err)
		}

		err = manager.Save(preset2)
		if err != nil {
			t.Fatalf("Failed to save preset2: %v", err)
		}

		// list them
		metadata, err := manager.List()
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(metadata) != 2 {
			t.Errorf("Expected 2 presets, got %d", len(metadata))
		}

		// verify contents
		names := make(map[string]bool)
		for _, meta := range metadata {
			names[meta.Name] = true

			if meta.SwayVersion != "1.11" {
				t.Errorf("Expected sway version 1.11, got %s", meta.SwayVersion)
			}

			if meta.WorkspaceCount != 1 {
				t.Errorf("Expected 1 workspace, got %d", meta.WorkspaceCount)
			}

			if meta.WindowCount != 1 {
				t.Errorf("Expected 1 window, got %d", meta.WindowCount)
			}
		}

		if !names["layout-1"] || !names["layout-2"] {
			t.Error("Expected to find both layout-1 and layout-2 in the list")
		}
	})
}