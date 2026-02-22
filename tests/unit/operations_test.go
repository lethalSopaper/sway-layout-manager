package unit

import (
	"testing"
	"github.com/lidsol/sway-layout-manager/internal/layout"
)

func createTestWorkspaces() []layout.WorkspaceLayout {
	return []layout.WorkspaceLayout{
		{
			Num: 1,
			Name: "1",
			Output: "HDMI-A-1",
			Layout: "splith",
			Containers: []layout.ContainerLayout{
				{
					ID: 1,
					Type: "con",
					AppID: "firefox",
					WindowClass: "Firefox",
					ExecCommand: "firefox",
				},
				{
					ID: 2,
					Type: "con",
					AppID: "code",
					WindowClass: "Code",
					ExecCommand: "code",
				},
			},
		},
		{
			Num: 2,
			Name: "2",
			Output: "HDMI-A-1",
			Layout: "splith",
			Containers: []layout.ContainerLayout{
				{
					ID: 3,
					Type: "con",
					AppID: "foot",
					ExecCommand: "foot",
				},
			},
		},
		{
			Num: -1,
			Name: "named-workspace",
			Output: "HDMI-A-1",
			Layout: "splith",
			Containers: []layout.ContainerLayout{
				{
					ID: 4,
					Type: "con",
					AppID: "thunar",
					ExecCommand: "thunar",
				},
			},
		},
	}
}

func TestFilterSkipWorkspaces(t *testing.T) {
	workspaces := createTestWorkspaces()

	t.Run("skip by number", func(t *testing.T) {
		filtered := layout.FilterSkipWorkspaces(workspaces, []string{"1"})

		if len(filtered) != 2 {
			t.Errorf("Expected 2 workspaces, got %d", len(filtered))
		}

		for _, ws := range filtered {
			if ws.Num == 1 {
				t.Error("Workspace 1 should be skipped")
			}
		}
	})

	t.Run("skip by name", func(t *testing.T) {
		filtered := layout.FilterSkipWorkspaces(workspaces, []string{"named-workspace"})

		if len(filtered) != 2 {
			t.Errorf("Expected 2 workspaces, got %d", len(filtered))
		}

		for _, ws := range filtered {
			if ws.Name == "named-workspace" {
				t.Error("named-workspace should be skipped")
			}
		}
	})

	t.Run("skip multiple workspaces", func(t *testing.T) {
		filtered := layout.FilterSkipWorkspaces(workspaces, []string{"1", "2"})

		if len(filtered) != 1 {
			t.Errorf("Expected 1 workspace, got %d", len(filtered))
		}

		if filtered[0].Num != -1 {
			t.Error("Only named-workspace should remain")
		}
	})

	t.Run("skip non-existent workspace", func(t *testing.T) {
		filtered := layout.FilterSkipWorkspaces(workspaces, []string{"999"})

		if len(filtered) != 3 {
			t.Errorf("Expected 3 workspaces (no change), got %d", len(filtered))
		}
	})

	t.Run("empty skip list", func(t *testing.T) {
		filtered := layout.FilterSkipWorkspaces(workspaces, []string{})

		if len(filtered) != 3 {
			t.Errorf("Expected 3 workspaces (no change), got %d", len(filtered))
		}
	})
}

func TestFilterOnlyWorkspaces(t *testing.T) {
	workspaces := createTestWorkspaces()

	t.Run("only by number", func(t *testing.T) {
		filtered := layout.FilterOnlyWorkspaces(workspaces, []string{"1"})

		if len(filtered) != 1 {
			t.Errorf("Expected 1 workspace, got %d", len(filtered))
		}

		if filtered[0].Num != 1 {
			t.Error("Only workspace 1 should be included")
		}
	})

	t.Run("only by name", func(t *testing.T) {
		filtered := layout.FilterOnlyWorkspaces(workspaces, []string{"named-workspace"})

		if len(filtered) != 1 {
			t.Errorf("Expected 1 workspace, got %d", len(filtered))
		}

		if filtered[0].Name != "named-workspace" {
			t.Error("Only named-workspace should be included")
		}
	})

	t.Run("only multiple workspaces", func(t *testing.T) {
		filtered := layout.FilterOnlyWorkspaces(workspaces, []string{"1", "2"})

		if len(filtered) != 2 {
			t.Errorf("Expected 2 workspaces, got %d", len(filtered))
		}
	})

	t.Run("empty only list", func(t *testing.T) {
		filtered := layout.FilterOnlyWorkspaces(workspaces, []string{})

		if len(filtered) != 3 {
			t.Errorf("Expected 3 workspaces (no change), got %d", len(filtered))
		}
	})
}

func TestFilterFloatingContainers(t *testing.T) {
	workspaces := []layout.WorkspaceLayout{
		{
			Num: 1,
			Name: "1",
			Output: "HDMI-A-1",
			Containers: []layout.ContainerLayout{
				{
					ID: 1,
					Type: "con",
					Floating: "auto_off",
					AppID: "firefox",
				},
				{
					ID: 2,
					Type: "floating_con",
					Floating: "user_on",
					AppID: "calculator",
				},
				{
					ID: 3,
					Type: "con",
					Floating: "auto_off",
					AppID: "code",
				},
			},
		},
	}

	filtered := layout.FilterFloatingContainers(workspaces)

	if len(filtered) != 1 {
		t.Fatalf("Expected 1 workspace, got %d", len(filtered))
	}

	if len(filtered[0].Containers) != 2 {
		t.Errorf("Expected 2 containers (non-floating), got %d", len(filtered[0].Containers))
	}

	for _, container := range filtered[0].Containers {
		if container.Floating == "user_on" {
			t.Error("Floating container should be filtered out")
		}
	}
}

func TestFilterSkipApps(t *testing.T) {
	workspaces := createTestWorkspaces()

	t.Run("skip by app_id", func(t *testing.T) {
		filtered := layout.FilterSkipApps(workspaces, []string{"firefox"})

		if len(filtered[0].Containers) != 1 {
			t.Errorf("Expected 1 container in workspace 1, got %d", len(filtered[0].Containers))
		}

		if filtered[0].Containers[0].AppID == "firefox" {
			t.Error("firefox should be skipped")
		}
	})

	t.Run("skip by class", func(t *testing.T) {
		filtered := layout.FilterSkipApps(workspaces, []string{"Firefox"})

		// check workspace 1
		if len(filtered[0].Containers) != 1 {
			t.Errorf("Expected 1 container in workspace 1, got %d", len(filtered[0].Containers))
		}

		if filtered[0].Containers[0].WindowClass == "Firefox" {
			t.Error("Firefox (by class) should be skipped")
		}
	})

	t.Run("skip multiple apps", func(t *testing.T) {
		filtered := layout.FilterSkipApps(workspaces, []string{"firefox", "code", "foot"})

		// workspace 1 should have no containers
		if len(filtered[0].Containers) != 0 {
			t.Errorf("Expected 0 containers in workspace 1, got %d", len(filtered[0].Containers))
		}

		// workspace 2 should be empty
		if len(filtered[1].Containers) != 0 {
			t.Errorf("Expected 0 containers in workspace 2, got %d", len(filtered[1].Containers))
		}

		// workspace 3 (named) should still have thunar
		if len(filtered[2].Containers) != 1 {
			t.Errorf("Expected 1 container in workspace 3, got %d", len(filtered[2].Containers))
		}
	})

	t.Run("empty skip list", func(t *testing.T) {
		filtered := layout.FilterSkipApps(workspaces, []string{})

		if len(filtered[0].Containers) != 2 {
			t.Error("No apps should be skipped with empty list")
		}
	})
}

func TestFilterOnlyApps(t *testing.T) {
	workspaces := createTestWorkspaces()

	t.Run("only by app_id", func(t *testing.T) {
		filtered := layout.FilterOnlyApps(workspaces, []string{"firefox"})

		// workspace 1 should have only firefox
		if len(filtered[0].Containers) != 1 {
			t.Errorf("Expected 1 container in workspace 1, got %d", len(filtered[0].Containers))
		}

		if filtered[0].Containers[0].AppID != "firefox" {
			t.Error("Only firefox should be included")
		}

		// workspace 2 should have no containers
		if len(filtered[1].Containers) != 0 {
			t.Errorf("Expected 0 containers in workspace 2, got %d", len(filtered[1].Containers))
		}
	})

	t.Run("only by class", func(t *testing.T) {
		filtered := layout.FilterOnlyApps(workspaces, []string{"Code"})

		// workspace 1 should have only code
		if len(filtered[0].Containers) != 1 {
			t.Errorf("Expected 1 container in workspace 1, got %d", len(filtered[0].Containers))
		}

		if filtered[0].Containers[0].WindowClass != "Code" {
			t.Error("Only Code should be included")
		}
	})

	t.Run("only multiple apps", func(t *testing.T) {
		filtered := layout.FilterOnlyApps(workspaces, []string{"firefox", "foot"})

		// workspace 1 should have only firefox
		if len(filtered[0].Containers) != 1 {
			t.Errorf("Expected 1 container in workspace 1, got %d", len(filtered[0].Containers))
		}

		// workspace 2 should have foot
		if len(filtered[1].Containers) != 1 {
			t.Errorf("Expected 1 container in workspace 2, got %d", len(filtered[1].Containers))
		}
	})

	t.Run("empty only list", func(t *testing.T) {
		filtered := layout.FilterOnlyApps(workspaces, []string{})

		if len(filtered[0].Containers) != 2 {
			t.Error("No filtering should occur with empty list")
		}
	})
}

func TestFilterNestedContainers(t *testing.T) {
	workspaces := []layout.WorkspaceLayout{
		{
			Num: 1,
			Name: "1",
			Containers: []layout.ContainerLayout{
				{
					ID: 1,
					Type: "con",
					Layout: "tabbed",
					Children: []layout.ContainerLayout{
						{
							ID: 2,
							Type: "con",
							AppID: "firefox",
							ExecCommand: "firefox",
						},
						{
							ID: 3,
							Type: "con",
							AppID: "chromium",
							ExecCommand: "chromium",
						},
					},
				},
			},
		},
	}

	t.Run("filter nested apps", func(t *testing.T) {
		filtered := layout.FilterSkipApps(workspaces, []string{"firefox"})

		// parent container should still exist
		if len(filtered[0].Containers) != 1 {
			t.Fatalf("Expected 1 parent container, got %d", len(filtered[0].Containers))
		}

		if len(filtered[0].Containers[0].Children) != 1 {
			t.Errorf("Expected 1 child container, got %d", len(filtered[0].Containers[0].Children))
		}

		if filtered[0].Containers[0].Children[0].AppID != "chromium" {
			t.Error("Only chromium should remain")
		}
	})
}
