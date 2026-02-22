package unit

import (
	"testing"
	"github.com/lidsol/sway-layout-manager/internal/layout"
	"github.com/lidsol/sway-layout-manager/internal/sway"
)

func TestNewRestorer(t *testing.T) {
	client, err := sway.NewClient()
	if err != nil {
		t.Skip("Skipping: swaymsg not available")
	}

	restorer := layout.NewRestorer(client)

	if restorer == nil {
		t.Fatal("NewRestorer() returned nil")
	}
}

func TestRestorerConfiguration(t *testing.T) {
	client, err := sway.NewClient()
	if err != nil {
		t.Skip("Skipping: swaymsg not available")
	}

	restorer := layout.NewRestorer(client)

	t.Run("default configuration", func(t *testing.T) {
		if restorer.ReuseExistingWindows {
			t.Error("ReuseExistingWindows should be false by default")
		}

		if len(restorer.ReuseApps) != 0 {
			t.Error("ReuseApps should be empty by default")
		}

		if restorer.ClearWorkspaces {
			t.Error("ClearWorkspaces should be false by default")
		}

		if restorer.ClearAllWorkspaces {
			t.Error("ClearAllWorkspaces should be false by default")
		}
	})

	t.Run("reuse existing windows configuration", func(t *testing.T) {
		restorer.ReuseExistingWindows = true
		if !restorer.ReuseExistingWindows {
			t.Error("Failed to set ReuseExistingWindows")
		}
	})

	t.Run("reuse specific apps configuration", func(t *testing.T) {
		restorer.ReuseApps = []string{"firefox", "code"}
		if len(restorer.ReuseApps) != 2 {
			t.Errorf("Expected 2 apps in ReuseApps, got %d", len(restorer.ReuseApps))
		}
	})

	t.Run("clear workspaces configuration", func(t *testing.T) {
		restorer.ClearWorkspaces = true
		if !restorer.ClearWorkspaces {
			t.Error("Failed to set ClearWorkspaces")
		}
	})

	t.Run("clear all workspaces configuration", func(t *testing.T) {
		restorer.ClearAllWorkspaces = true
		if !restorer.ClearAllWorkspaces {
			t.Error("Failed to set ClearAllWorkspaces")
		}
	})
}

func TestRestoreValidation(t *testing.T) {
	client, err := sway.NewClient()
	if err != nil {
		t.Skip("Skipping: swaymsg not available")
	}

	restorer := layout.NewRestorer(client)

	t.Run("restore with nil preset", func(t *testing.T) {
		result, err := restorer.Restore(nil)

		if err == nil {
			t.Error("Expected error when restoring nil preset")
		}

		if result != nil {
			t.Error("Expected nil result when restoring nil preset")
		}

		expectedMsg := "preset cannot be nil"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("restore with empty preset", func(t *testing.T) {
		preset := &layout.Preset{
			Name:       "empty-preset",
			Workspaces: []layout.WorkspaceLayout{},
		}

		result, err := restorer.Restore(preset)

		if err != nil {
			t.Errorf("Unexpected error with empty preset: %v", err)
		}

		if result == nil {
			t.Fatal("Expected non-nil result")
		}

		if result.WorkspacesRestored != 0 {
			t.Errorf("Expected 0 workspaces restored, got %d", result.WorkspacesRestored)
		}
	})
}

func TestRestoreResultStructure(t *testing.T) {
	result := &layout.RestoreResult{
		WorkspacesRestored: 2,
		ApplicationsLaunched: 5,
		ApplicationsReused: 3,
		ApplicationsFailed: 1,
		Errors: []error{},
	}

	if result.WorkspacesRestored != 2 {
		t.Errorf("Expected 2 workspaces restored, got %d", result.WorkspacesRestored)
	}

	if result.ApplicationsLaunched != 5 {
		t.Errorf("Expected 5 applications launched, got %d", result.ApplicationsLaunched)
	}

	if result.ApplicationsReused != 3 {
		t.Errorf("Expected 3 applications reused, got %d", result.ApplicationsReused)
	}

	if result.ApplicationsFailed != 1 {
		t.Errorf("Expected 1 application failed, got %d", result.ApplicationsFailed)
	}
	if len(result.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(result.Errors))
	}
}
