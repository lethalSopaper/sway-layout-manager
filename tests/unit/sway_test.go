package unit

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/lidsol/sway-layout-manager/internal/sway"
)

func TestNewClient(t *testing.T) {
	t.Run("successful client creation when swaymsg exists", func(t *testing.T) {
		// skip test if swaymsg is not available
		if _, err := exec.LookPath("swaymsg"); err != nil {
			t.Skip("swaymsg not available in PATH, skipping test")
		}

		client, err := sway.NewClient()
		if err != nil {
			t.Fatalf("Expected successful client creation, got error: %v", err)
		}

		if client == nil {
			t.Fatal("Expected non-nil client")
		}
	})

	t.Run("client creation fails when swaymsg not available", func(t *testing.T) {
		// modify PATH to simulate swaymsg not being available
		originalPath := os.Getenv("PATH")
		defer os.Setenv("PATH", originalPath)
		os.Setenv("PATH", "")

		client, err := sway.NewClient()
		if err == nil {
			t.Fatal("Expected error when swaymsg not available, got nil")
		}

		if client != nil {
			t.Fatal("Expected nil client when swaymsg not available")
		}

		expectedErrorMsg := "swaymsg not found in PATH"
		if !strings.Contains(err.Error(), expectedErrorMsg) {
			t.Errorf("Expected error message to contain '%s', got: %v", expectedErrorMsg, err)
		}
	})
}

func TestClientMethods(t *testing.T) {
	// skip client method tests if swaymsg is not available
	if _, err := exec.LookPath("swaymsg"); err != nil {
		t.Skip("swaymsg not available in PATH, skipping sway client method tests")
	}

	client, err := sway.NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	t.Run("GetTree method", func(t *testing.T) {
		tree, err := client.GetTree()
		if err != nil {
			t.Logf("GetTree failed (expected in non-Sway environment): %v", err)
			return
		}

		if tree == nil {
			t.Error("Expected non-nil tree")
		}

		// validation for the tree structure
		if tree.Type == "" {
			t.Error("Expected tree to have a type")
		}
	})

	t.Run("GetWorkspaces method", func(t *testing.T) {
		workspaces, err := client.GetWorkspaces()
		if err != nil {
			t.Logf("GetWorkspaces failed (expected in non-Sway environment): %v", err)
			return
		}

		// workspaces can be empty
		if workspaces == nil {
			t.Error("Expected non-nil workspaces slice")
		}
	})

	t.Run("GetVersion method", func(t *testing.T) {
		version, err := client.GetVersion()
		if err != nil {
			t.Logf("GetVersion failed (expected in non-Sway environment): %v", err)
			return
		}

		if version == nil {
			t.Error("Expected non-nil version")
		}

		if version.HumanReadable == "" {
			t.Error("Expected version to have human readable string")
		}
	})
}

func TestSwayTypes(t *testing.T) {
	t.Run("Node JSON unmarshaling", func(t *testing.T) {
		nodeJSON := `{
			"id": 123,
			"name": "test-node",
			"type": "con",
			"layout": "splith",
			"orientation": "horizontal",
			"urgent": false,
			"focused": true,
			"nodes": [],
			"floating_nodes": [],
			"sticky": false,
			"idle_inhibitors": {}
		}`

		var node sway.Node
		err := json.Unmarshal([]byte(nodeJSON), &node)
		if err != nil {
			t.Fatalf("Failed to unmarshal node JSON: %v", err)
		}

		// validate parsed values
		if node.ID != 123 {
			t.Errorf("Expected ID 123, got %d", node.ID)
		}
		if node.Name != "test-node" {
			t.Errorf("Expected name 'test-node', got '%s'", node.Name)
		}
		if node.Type != "con" {
			t.Errorf("Expected type 'con', got '%s'", node.Type)
		}
		if node.Layout != "splith" {
			t.Errorf("Expected layout 'splith', got '%s'", node.Layout)
		}
		if !node.Focused {
			t.Error("Expected focused to be true")
		}
	})

	t.Run("Workspace JSON unmarshaling", func(t *testing.T) {
		workspaceJSON := `{
			"num": 1,
			"name": "1",
			"visible": true,
			"focused": false,
			"urgent": false,
			"layout": "splith",
			"orientation": "horizontal"
		}`

		var workspace sway.Workspace
		err := json.Unmarshal([]byte(workspaceJSON), &workspace)
		if err != nil {
			t.Fatalf("Failed to unmarshal workspace JSON: %v", err)
		}

		// Validate parsed values
		if workspace.Num != 1 {
			t.Errorf("Expected Num 1, got %d", workspace.Num)
		}
		if workspace.Name != "1" {
			t.Errorf("Expected name '1', got '%s'", workspace.Name)
		}
		if !workspace.Visible {
			t.Error("Expected visible to be true")
		}
		if workspace.Focused {
			t.Error("Expected focused to be false")
		}
	})

	t.Run("Version JSON unmarshaling", func(t *testing.T) {
		versionJSON := `{
			"human_readable": "sway version 1.8.1",
			"variant": "sway",
			"major": 1,
			"minor": 8,
			"patch": 1,
			"loaded_config_file_name": "/home/user/.config/sway/config"
		}`

		var version sway.Version
		err := json.Unmarshal([]byte(versionJSON), &version)
		if err != nil {
			t.Fatalf("Failed to unmarshal version JSON: %v", err)
		}

		// validate parsed values
		if version.HumanReadable != "sway version 1.8.1" {
			t.Errorf("Expected human readable 'sway version 1.8.1', got '%s'", version.HumanReadable)
		}
		if version.Variant != "sway" {
			t.Errorf("Expected variant 'sway', got '%s'", version.Variant)
		}
		if version.Major != 1 || version.Minor != 8 || version.Patch != 1 {
			t.Errorf("Expected version 1.8.1, got %d.%d.%d", version.Major, version.Minor, version.Patch)
		}
	})
}