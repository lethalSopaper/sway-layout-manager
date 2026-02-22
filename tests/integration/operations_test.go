package integration

import (
	"testing"
	"time"
	"github.com/lidsol/sway-layout-manager/internal/config"
	"github.com/lidsol/sway-layout-manager/internal/layout"
	"github.com/lidsol/sway-layout-manager/internal/sway"
	"github.com/lidsol/sway-layout-manager/pkg/preset"
)

func TestSaveWithWorkspaceFiltering(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	client, err := sway.NewClient()
	if err != nil {
		t.Skip("Skipping: swaymsg not available")
	}

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	parser := layout.NewParser(client)
	manager := preset.NewManager(cfg)

	t.Run("save and skip specific workspace", func(t *testing.T) {
		// capture current layout
		captured, err := parser.CaptureCurrentLayout("test-skip-workspace")
		if err != nil {
			t.Fatalf("Failed to capture layout: %v", err)
		}

		originalCount := len(captured.Workspaces)

		captured.Workspaces = layout.FilterSkipWorkspaces(captured.Workspaces, []string{"1"})

		if len(captured.Workspaces) >= originalCount {
			t.Skip("Workspace 1 doesn't exist, skipping test")
		}

		// save filtered preset
		err = manager.Save(captured)
		if err != nil {
			t.Fatalf("Failed to save preset: %v", err)
		}

		// reload and verify
		loaded, err := manager.Load("test-skip-workspace")
		if err != nil {
			t.Fatalf("Failed to load preset: %v", err)
		}

		// verify workspace 1 is not in the saved preset
		for _, ws := range loaded.Workspaces {
			if ws.Num == 1 {
				t.Error("Workspace 1 should have been filtered out")
			}
		}
	})

	t.Run("save only specific workspace", func(t *testing.T) {
		captured, err := parser.CaptureCurrentLayout("test-only-workspace")
		if err != nil {
			t.Fatalf("Failed to capture layout: %v", err)
		}

		// filter to only workspace 1 (if it exists)
		captured.Workspaces = layout.FilterOnlyWorkspaces(captured.Workspaces, []string{"1"})

		err = manager.Save(captured)
		if err != nil {
			t.Fatalf("Failed to save preset: %v", err)
		}

		loaded, err := manager.Load("test-only-workspace")
		if err != nil {
			t.Fatalf("Failed to load preset: %v", err)
		}

		// all workspaces should have num=1 or name="1"
		for _, ws := range loaded.Workspaces {
			if ws.Num != 1 && ws.Name != "1" {
				t.Errorf("Only workspace 1 should be saved, found workspace %d (%s)", ws.Num, ws.Name)
			}
		}
	})
}

func TestSaveWithAppFiltering(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	client, err := sway.NewClient()
	if err != nil {
		t.Skip("Skipping: swaymsg not available")
	}

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	parser := layout.NewParser(client)
	manager := preset.NewManager(cfg)

	t.Run("save and skip specific app", func(t *testing.T) {
		captured, err := parser.CaptureCurrentLayout("test-skip-app")
		if err != nil {
			t.Fatalf("Failed to capture layout: %v", err)
		}

		// count containers before filtering
		originalCount := 0
		for _, ws := range captured.Workspaces {
			originalCount += countContainersInWorkspace(ws)
		}

		// filter out firefox (if it exists)
		captured.Workspaces = layout.FilterSkipApps(captured.Workspaces, []string{"firefox", "Firefox"})

		newCount := 0
		for _, ws := range captured.Workspaces {
			newCount += countContainersInWorkspace(ws)
		}

		if newCount > originalCount {
			t.Error("Container count should not increase after filtering")
		}

		t.Logf("Filtered from %d to %d containers", originalCount, newCount)

		err = manager.Save(captured)
		if err != nil {
			t.Fatalf("Failed to save preset: %v", err)
		}

		loaded, err := manager.Load("test-skip-app")
		if err != nil {
			t.Fatalf("Failed to load preset: %v", err)
		}

		// verify firefox is not in any container
		for _, ws := range loaded.Workspaces {
			verifyAppNotInContainers(t, ws.Containers, "firefox", "Firefox")
		}
	})

	t.Run("save only specific app", func(t *testing.T) {
		captured, err := parser.CaptureCurrentLayout("test-only-app")
		if err != nil {
			t.Fatalf("Failed to capture layout: %v", err)
		}

		// filter to only keep code editor (if it exists)
		captured.Workspaces = layout.FilterOnlyApps(captured.Workspaces, []string{"code", "Code"})

		err = manager.Save(captured)
		if err != nil {
			t.Fatalf("Failed to save preset: %v", err)
		}

		loaded, err := manager.Load("test-only-app")
		if err != nil {
			t.Fatalf("Failed to load preset: %v", err)
		}

		// all containers should be code or have code in children
		for _, ws := range loaded.Workspaces {
			for _, container := range ws.Containers {
				if !containsApp(container, "code", "Code") {
					t.Logf("Container %d doesn't contain code (might be parent container)", container.ID)
				}
			}
		}
	})
}

func TestSaveWithFloatingFilter(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	client, err := sway.NewClient()
	if err != nil {
		t.Skip("Skipping: swaymsg not available")
	}

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	parser := layout.NewParser(client)
	manager := preset.NewManager(cfg)

	t.Run("filter floating windows", func(t *testing.T) {
		captured, err := parser.CaptureCurrentLayout("test-no-floating")
		if err != nil {
			t.Fatalf("Failed to capture layout: %v", err)
		}

		// filter out floating windows
		captured.Workspaces = layout.FilterFloatingContainers(captured.Workspaces)

		err = manager.Save(captured)
		if err != nil {
			t.Fatalf("Failed to save preset: %v", err)
		}

		loaded, err := manager.Load("test-no-floating")
		if err != nil {
			t.Fatalf("Failed to load preset: %v", err)
		}

		// verify no floating containers
		for _, ws := range loaded.Workspaces {
			verifyNoFloatingContainers(t, ws.Containers)
		}
	})
}

func TestListWithFilteredPresets(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	manager := preset.NewManager(cfg)

	// create presets with different configurations
	presets := []*layout.Preset{
		{
			Name: "preset-1-workspace",
			CreatedAt: time.Now(),
			SwayVersion: "1.11",
			Workspaces: []layout.WorkspaceLayout{
				{Num: 1, Name: "1", Containers: []layout.ContainerLayout{}},
			},
		},
		{
			Name: "preset-2-workspaces",
			CreatedAt: time.Now(),
			SwayVersion: "1.11",
			Workspaces: []layout.WorkspaceLayout{
				{Num: 1, Name: "1", Containers: []layout.ContainerLayout{}},
				{Num: 2, Name: "2", Containers: []layout.ContainerLayout{}},
			},
		},
		{
			Name:"preset-with-apps",
			CreatedAt: time.Now(),
			SwayVersion: "1.11",
			Workspaces: []layout.WorkspaceLayout{
				{
					Num: 1,
					Name: "1",
					Containers: []layout.ContainerLayout{
						{ID: 1, AppID: "firefox", ExecCommand: "firefox"},
						{ID: 2, AppID: "code", ExecCommand: "code"},
					},
				},
			},
		},
	}

	for _, p := range presets {
		err := manager.Save(p)
		if err != nil {
			t.Fatalf("Failed to save preset %s: %v", p.Name, err)
		}
	}

	metadata, err := manager.List()
	if err != nil {
		t.Fatalf("Failed to list presets: %v", err)
	}

	if len(metadata) < 3 {
		t.Errorf("Expected at least 3 presets, got %d", len(metadata))
	}

	// verify metadata contains correct information
	for _, meta := range metadata {
		if meta.WorkspaceCount < 0 {
			t.Errorf("Invalid workspace count for %s: %d", meta.Name, meta.WorkspaceCount)
		}

		if meta.WindowCount < 0 {
			t.Errorf("Invalid window count for %s: %d", meta.Name, meta.WindowCount)
		}

		if meta.SwayVersion == "" {
			t.Errorf("Missing SwayVersion for %s", meta.Name)
		}
	}
}

func TestDeleteFilteredPresets(t *testing.T) {
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	manager := preset.NewManager(cfg)

	// create and save a preset
	preset := &layout.Preset{
		Name:        "test-delete-filtered",
		CreatedAt:   time.Now(),
		SwayVersion: "1.11",
		Workspaces: []layout.WorkspaceLayout{
			{Num: 1, Name: "1", Containers: []layout.ContainerLayout{}},
		},
	}

	err = manager.Save(preset)
	if err != nil {
		t.Fatalf("Failed to save preset: %v", err)
	}

	// verify it exists
	if !manager.Exists("test-delete-filtered") {
		t.Fatal("Preset should exist after saving")
	}

	// delete it
	err = manager.Delete("test-delete-filtered")
	if err != nil {
		t.Fatalf("Failed to delete preset: %v", err)
	}

	// verify it's gone
	if manager.Exists("test-delete-filtered") {
		t.Error("Preset should not exist after deletion")
	}
}

func countContainersInWorkspace(ws layout.WorkspaceLayout) int {
	count := 0
	for _, container := range ws.Containers {
		count += countContainersRecursive(container)
	}
	return count
}

func countContainersRecursive(container layout.ContainerLayout) int {
	count := 1
	for _, child := range container.Children {
		count += countContainersRecursive(child)
	}
	return count
}

func verifyAppNotInContainers(t *testing.T, containers []layout.ContainerLayout, appID, class string) {
	for _, container := range containers {
		if container.AppID == appID || container.WindowClass == class {
			t.Errorf("Found filtered app: %s/%s", appID, class)
		}
		if len(container.Children) > 0 {
			verifyAppNotInContainers(t, container.Children, appID, class)
		}
	}
}

func containsApp(container layout.ContainerLayout, appID, class string) bool {
	if container.AppID == appID || container.WindowClass == class {
		return true
	}
	for _, child := range container.Children {
		if containsApp(child, appID, class) {
			return true
		}
	}
	return false
}

func verifyNoFloatingContainers(t *testing.T, containers []layout.ContainerLayout) {
	for _, container := range containers {
		if container.Floating == "user_on" || container.Type == "floating_con" {
			t.Errorf("Found floating container: ID=%d, Floating=%s, Type=%s",
				container.ID, container.Floating, container.Type)
		}
		if len(container.Children) > 0 {
			verifyNoFloatingContainers(t, container.Children)
		}
	}
}
