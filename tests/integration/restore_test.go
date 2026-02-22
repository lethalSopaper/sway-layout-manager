package integration

import (
	"testing"
	"time"
	"github.com/lidsol/sway-layout-manager/internal/layout"
	"github.com/lidsol/sway-layout-manager/internal/sway"
)

func TestRestoreBasicFunctionality(t *testing.T) {
	client, err := sway.NewClient()
	if err != nil {
		t.Skip("Skipping: swaymsg not available")
	}

	restorer := layout.NewRestorer(client)

	t.Run("restore empty preset", func(t *testing.T) {
		preset := &layout.Preset{
			Name: "empty-test",
			CreatedAt: time.Now(),
			SwayVersion: "1.11",
			Outputs: []layout.OutputLayout{},
			Workspaces: []layout.WorkspaceLayout{},
		}

		result, err := restorer.Restore(preset)
		if err != nil {
			t.Fatalf("Failed to restore empty preset: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		if result.WorkspacesRestored != 0 {
			t.Errorf("Expected 0 workspaces restored, got %d", result.WorkspacesRestored)
		}

		if result.ApplicationsLaunched != 0 {
			t.Errorf("Expected 0 applications launched, got %d", result.ApplicationsLaunched)
		}
	})
}

func TestRestoreWithReuseConfiguration(t *testing.T) {
	client, err := sway.NewClient()
	if err != nil {
		t.Skip("Skipping: swaymsg not available")
	}

	t.Run("configure reuse existing windows", func(t *testing.T) {
		restorer := layout.NewRestorer(client)
		restorer.ReuseExistingWindows = true

		if !restorer.ReuseExistingWindows {
			t.Error("Failed to configure ReuseExistingWindows")
		}

		preset := &layout.Preset{
			Name: "reuse-test",
			CreatedAt: time.Now(),
			SwayVersion: "1.11",
			Workspaces: []layout.WorkspaceLayout{},
		}

		_, err := restorer.Restore(preset)
		if err != nil {
			t.Errorf("Restore failed with ReuseExistingWindows: %v", err)
		}
	})

	t.Run("configure reuse specific apps", func(t *testing.T) {
		restorer := layout.NewRestorer(client)
		restorer.ReuseApps = []string{"firefox", "code"}

		if len(restorer.ReuseApps) != 2 {
			t.Error("Failed to configure ReuseApps")
		}

		preset := &layout.Preset{
			Name:        "reuse-apps-test",
			CreatedAt:   time.Now(),
			SwayVersion: "1.11",
			Workspaces:  []layout.WorkspaceLayout{},
		}

		_, err := restorer.Restore(preset)
		if err != nil {
			t.Errorf("Restore failed with ReuseApps: %v", err)
		}
	})
}

func TestRestoreWithClearConfiguration(t *testing.T) {
	client, err := sway.NewClient()
	if err != nil {
		t.Skip("Skipping: swaymsg not available")
	}

	t.Run("configure clear workspaces", func(t *testing.T) {
		restorer := layout.NewRestorer(client)
		restorer.ClearWorkspaces = true

		preset := &layout.Preset{
			Name: "clear-workspaces-test",
			CreatedAt: time.Now(),
			SwayVersion: "1.11",
			Workspaces: []layout.WorkspaceLayout{
				{
					Num: 98,
					Name: "98",
					Output: "HDMI-A-1",
					Layout: "splith",
					Containers: []layout.ContainerLayout{},
				},
			},
		}

		_, err := restorer.Restore(preset)
		if err != nil {
			t.Errorf("Restore failed with ClearWorkspaces: %v", err)
		}
	})

	t.Run("configure clear all workspaces", func(t *testing.T) {
		restorer := layout.NewRestorer(client)
		restorer.ClearAllWorkspaces = true

		preset := &layout.Preset{
			Name: "clear-all-test",
			CreatedAt: time.Now(),
			SwayVersion: "1.11",
			Workspaces: []layout.WorkspaceLayout{
				{
					Num: 97,
					Name: "97",
					Output: "HDMI-A-1",
					Layout: "splith",
					Containers: []layout.ContainerLayout{},
				},
			},
		}

		_, err := restorer.Restore(preset)
		if err != nil {
			t.Errorf("Restore failed with ClearAllWorkspaces: %v", err)
		}
	})
}

func TestRestoreErrorHandling(t *testing.T) {
	client, err := sway.NewClient()
	if err != nil {
		t.Skip("Skipping: swaymsg not available")
	}

	restorer := layout.NewRestorer(client)

	t.Run("restore nil preset", func(t *testing.T) {
		_, err := restorer.Restore(nil)
		if err == nil {
			t.Error("Expected error when restoring nil preset")
		}
	})

	t.Run("restore preset with invalid workspace", func(t *testing.T) {
		preset := &layout.Preset{
			Name: "invalid-workspace-test",
			CreatedAt: time.Now(),
			SwayVersion: "1.11",
			Workspaces: []layout.WorkspaceLayout{
				{
					Num: -2,
					Name: "",
					Output: "INVALID-OUTPUT",
					Layout: "splith",
					Containers: []layout.ContainerLayout{},
				},
			},
		}

		result, err := restorer.Restore(preset)
		if err != nil {
			t.Logf("Restore returned error (expected): %v", err)
		}

		if result != nil {
			t.Logf("Restored with result: workspaces=%d, errors=%d",
				result.WorkspacesRestored, len(result.Errors))
		}
	})
}

func TestRestoreResultTracking(t *testing.T) {
	client, err := sway.NewClient()
	if err != nil {
		t.Skip("Skipping: swaymsg not available")
	}

	restorer := layout.NewRestorer(client)

	t.Run("result tracks workspaces", func(t *testing.T) {
		preset := &layout.Preset{
			Name: "tracking-test",
			CreatedAt: time.Now(),
			SwayVersion: "1.11",
			Workspaces: []layout.WorkspaceLayout{
				{
					Num: 96,
					Name: "96",
					Output: "HDMI-A-1",
					Layout: "splith",
					Containers: []layout.ContainerLayout{},
				},
				{
					Num: 95,
					Name: "95",
					Output: "HDMI-A-1",
					Layout: "splith",
					Containers: []layout.ContainerLayout{},
				},
			},
		}

		result, err := restorer.Restore(preset)
		if err != nil {
			t.Fatalf("Failed to restore: %v", err)
		}

		if result.WorkspacesRestored != 2 {
			t.Errorf("Expected 2 workspaces restored, got %d", result.WorkspacesRestored)
		}
	})

	t.Run("result structure is complete", func(t *testing.T) {
		preset := &layout.Preset{
			Name: "result-structure-test",
			CreatedAt: time.Now(),
			SwayVersion: "1.11",
			Workspaces: []layout.WorkspaceLayout{},
		}

		result, err := restorer.Restore(preset)
		if err != nil {
			t.Fatalf("Failed to restore: %v", err)
		}

		if result.WorkspacesRestored < 0 {
			t.Error("WorkspacesRestored should be non-negative")
		}
		if result.ApplicationsLaunched < 0 {
			t.Error("ApplicationsLaunched should be non-negative")
		}
		if result.ApplicationsReused < 0 {
			t.Error("ApplicationsReused should be non-negative")
		}
		if result.ApplicationsFailed < 0 {
			t.Error("ApplicationsFailed should be non-negative")
		}
		if result.Errors == nil {
			t.Error("Errors slice should be initialized")
		}
	})
}