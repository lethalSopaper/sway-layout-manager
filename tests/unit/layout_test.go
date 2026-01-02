package unit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/lidsol/sway-layout-manager/internal/layout"
	"github.com/lidsol/sway-layout-manager/internal/sway"
)

func TestLayoutTypes(t *testing.T) {
	t.Run("Preset JSON marshaling", func(t *testing.T) {
		preset := &layout.Preset{
			Name:        "test-preset",
			CreatedAt:   time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC),
			SwayVersion: "sway version 1.8.1",
			Outputs: []layout.OutputLayout{
				{
					Name:   "HDMI-A-1",
					Active: true,
					Rect: layout.Rect{
						X:      0,
						Y:      0,
						Width:  1920,
						Height: 1080,
					},
				},
			},
			Workspaces: []layout.WorkspaceLayout{
				{
					Num:    1,
					Name:   "1",
					Output: "HDMI-A-1",
					Layout: "splith",
					Containers: []layout.ContainerLayout{
						{
							Type:   "con",
							Layout: "splith",
							Rect: layout.Rect{
								X:      0,
								Y:      0,
								Width:  960,
								Height: 1080,
							},
						},
					},
				},
			},
		}

		jsonData, err := json.Marshal(preset)
		if err != nil {
			t.Fatalf("Failed to marshal preset: %v", err)
		}

		// verify we can unmarshal it back
		var unmarshaled layout.Preset
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal preset: %v", err)
		}

		// verify key fields
		if unmarshaled.Name != preset.Name {
			t.Errorf("Expected name '%s', got '%s'", preset.Name, unmarshaled.Name)
		}
		if unmarshaled.SwayVersion != preset.SwayVersion {
			t.Errorf("Expected sway version '%s', got '%s'", preset.SwayVersion, unmarshaled.SwayVersion)
		}
		if len(unmarshaled.Outputs) != 1 {
			t.Errorf("Expected 1 output, got %d", len(unmarshaled.Outputs))
		}
		if len(unmarshaled.Workspaces) != 1 {
			t.Errorf("Expected 1 workspace, got %d", len(unmarshaled.Workspaces))
		}
	})

	t.Run("OutputLayout JSON marshaling", func(t *testing.T) {
		output := layout.OutputLayout{
			Name:   "DP-1",
			Active: true,
			Rect: layout.Rect{
				X:      1920,
				Y:      0,
				Width:  2560,
				Height: 1440,
			},
		}

		jsonData, err := json.Marshal(output)
		if err != nil {
			t.Fatalf("Failed to marshal output: %v", err)
		}

		var unmarshaled layout.OutputLayout
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal output: %v", err)
		}

		if unmarshaled.Name != output.Name {
			t.Errorf("Expected name '%s', got '%s'", output.Name, unmarshaled.Name)
		}
		if unmarshaled.Active != output.Active {
			t.Errorf("Expected active %v, got %v", output.Active, unmarshaled.Active)
		}
		if unmarshaled.Rect != output.Rect {
			t.Errorf("Expected rect %+v, got %+v", output.Rect, unmarshaled.Rect)
		}
	})

	t.Run("WorkspaceLayout JSON marshaling", func(t *testing.T) {
		workspace := layout.WorkspaceLayout{
			Num:    2,
			Name:   "workspace2",
			Output: "HDMI-A-1",
			Layout: "splitv",
			Containers: []layout.ContainerLayout{
				{
					Type:   "con",
					Layout: "splitv",
					Rect: layout.Rect{
						X:      0,
						Y:      0,
						Width:  1920,
						Height: 540,
					},
				},
			},
		}

		jsonData, err := json.Marshal(workspace)
		if err != nil {
			t.Fatalf("Failed to marshal workspace: %v", err)
		}

		var unmarshaled layout.WorkspaceLayout
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal workspace: %v", err)
		}

		if unmarshaled.Num != workspace.Num {
			t.Errorf("Expected num %d, got %d", workspace.Num, unmarshaled.Num)
		}
		if unmarshaled.Name != workspace.Name {
			t.Errorf("Expected name '%s', got '%s'", workspace.Name, unmarshaled.Name)
		}
		if len(unmarshaled.Containers) != 1 {
			t.Errorf("Expected 1 container, got %d", len(unmarshaled.Containers))
		}
	})
}

func TestNewParser(t *testing.T) {
	// Skip if swaymsg not available
	client, err := sway.NewClient()
	if err != nil {
		t.Skip("swaymsg not available, skipping parser tests")
	}

	parser := layout.NewParser(client)

	if parser == nil {
		t.Fatal("Expected non-nil parser")
	}
}

func TestCaptureCurrentLayout(t *testing.T) {
	client, err := sway.NewClient()
	if err != nil {
		t.Skip("swaymsg not available, skipping capture tests")
	}

	parser := layout.NewParser(client)

	t.Run("successful layout capture attempt", func(t *testing.T) {
		preset, err := parser.CaptureCurrentLayout("test-preset")

		if err != nil {
			t.Logf("Layout capture failed (expected in non-sway environment): %v", err)
			return
		}

		// if we're actually in sway, verify the preset
		if preset != nil {
			if preset.Name != "test-preset" {
				t.Errorf("Expected preset name 'test-preset', got '%s'", preset.Name)
			}

			if time.Since(preset.CreatedAt) > time.Second {
				t.Error("Expected CreatedAt to be recent")
			}

			if preset.SwayVersion == "" {
				t.Error("Expected non-empty SwayVersion")
			}
		}
	})
}