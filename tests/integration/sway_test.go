package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/lidsol/sway-layout-manager/internal/sway"
)

// tests detection of sway environment
func TestSwayAvailability(t *testing.T) {
	t.Run("swaymsg command availability", func(t *testing.T) {
		// check if swaymsg is available in PATH
		_, err := exec.LookPath("swaymsg")

		if err != nil {
			t.Skip("swaymsg not available in PATH, skipping sway environment tests")
		}

		// creates a sway client
		client, err := sway.NewClient()
		if err != nil {
			t.Errorf("Failed to create sway client even though swaymsg is available: %v", err)
		}

		if client == nil {
			t.Error("Client should not be nil when swaymsg is available")
		}

		t.Log("Sway client created successfully")
	})

	t.Run("sway client creation without swaymsg", func(t *testing.T) {
		// temporarily remove swaymsg from PATH
		originalPath := getEnv("PATH")
		defer setEnv("PATH", originalPath)
		setEnv("PATH", "/nonexistent")

		client, err := sway.NewClient()
		if err == nil {
			t.Error("Client creation should fail when swaymsg is not available")
		}

		if client != nil {
			t.Error("Client should be nil when swaymsg is not available")
		}

		if !strings.Contains(err.Error(), "swaymsg") {
			t.Errorf("Error should mention swaymsg: %v", err)
		}

		t.Logf("Correctly failed to create client without swaymsg: %v", err)
	})
}

// tests different sway connection scenarios
func TestSwayConnectionAttempts(t *testing.T) {
	// skik tests if swaymsg not available
	if _, err := exec.LookPath("swaymsg"); err != nil {
		t.Skip("swaymsg not available, skipping sway connection tests")
	}

	client, err := sway.NewClient()
	if err != nil {
		t.Fatalf("Failed to create sway client: %v", err)
	}

	t.Run("get sway version", func(t *testing.T) {
		version, err := client.GetVersion()

		if err != nil {
			t.Logf("GetVersion failed (expected in non-sway environment): %v", err)

			if !strings.Contains(strings.ToLower(err.Error()), "sway") &&
			   !strings.Contains(strings.ToLower(err.Error()), "connect") &&
			   !strings.Contains(strings.ToLower(err.Error()), "ipc") {
				t.Errorf("Error message should be informative about sway connection: %v", err)
			}
			return
		}

		// if we're actually in sway, validate the response
		if version == nil {
			t.Error("Version should not be nil on successful call")
		} else {
			if version.HumanReadable == "" {
				t.Error("Version should have human readable string")
			}

			if version.Variant == "" {
				t.Error("Version should have variant field")
			}

			if version.Major == 0 && version.Minor == 0 && version.Patch == 0 {
				t.Error("Version should have numeric version components")
			}

			t.Logf("Sway version: %s (v%d.%d.%d)",
				version.HumanReadable, version.Major, version.Minor, version.Patch)
		}
	})

	t.Run("get sway tree", func(t *testing.T) {
		tree, err := client.GetTree()

		if err != nil {
			t.Logf("GetTree failed (expected in non-sway environment): %v", err)
			return
		}

		if tree == nil {
			t.Error("Tree should not be nil on successful call")
		} else {
			// basic tree validation
			if tree.Type == "" {
				t.Error("Root node should have a type")
			}

			if tree.Type != "root" {
				t.Errorf("Root node type should be 'root', got '%s'", tree.Type)
			}

			// tree should have some structure
			if len(tree.Nodes) == 0 {
				t.Log("Warning: Tree has no child nodes (might be empty sway session)")
			}

			t.Logf("Sway tree retrieved successfully: type=%s, children=%d", 
				tree.Type, len(tree.Nodes))
		}
	})

	t.Run("get sway workspaces", func(t *testing.T) {
		workspaces, err := client.GetWorkspaces()

		if err != nil {
			t.Logf("GetWorkspaces failed (expected in non-sway environment): %v", err)
			return
		}

		if workspaces == nil {
			t.Error("Workspaces should not be nil on successful call")
		} else {
			// workspaces can be empty in a fresh sway session
			t.Logf("Retrieved %d workspaces from sway", len(workspaces))

			for i, ws := range workspaces {
				if ws.Name == "" {
					t.Errorf("Workspace %d should have a name", i)
				}
				// workspace numbers can be negative in sway (that's valid)
				t.Logf("Workspace %d: name='%s', num=%d, visible=%v, focused=%v", 
					i, ws.Name, ws.Num, ws.Visible, ws.Focused)
			}
		}
	})
}

// tests parsing of actual sway JSON responses
func TestSwayJSONParsing(t *testing.T) {
	t.Run("parse sway version JSON", func(t *testing.T) {
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
			t.Fatalf("Failed to parse version JSON: %v", err)
		}

		// validate parsed fields
		expectedValues := map[string]interface{}{
			"human_readable": "sway version 1.8.1",
			"variant": "sway",
			"major": 1,
			"minor": 8,
			"patch": 1,
			"loaded_config_file_name": "/home/user/.config/sway/config",
		}
		
		actualValues := map[string]interface{}{
			"human_readable": version.HumanReadable,
			"variant": version.Variant,
			"major": version.Major,
			"minor": version.Minor,
			"patch": version.Patch,
			"loaded_config_file_name": version.LoadedConfigFileName,
		}
		
		for field, expected := range expectedValues {
			if actual := actualValues[field]; actual != expected {
				t.Errorf("Version field %s: expected %v, got %v", field, expected, actual)
			}
		}
	})
	
	t.Run("parse sway workspace JSON", func(t *testing.T) {
		workspaceJSON := `[
			{
				"id": 1,
				"num": 1,
				"name": "1",
				"visible": true,
				"focused": true,
				"urgent": false,
				"rect": {
					"x": 0,
					"y": 0,
					"width": 1920,
					"height": 1080
				},
				"output": "HDMI-A-1"
			},
			{
				"id": 2,
				"num": 2,
				"name": "web",
				"visible": false,
				"focused": false,
				"urgent": false,
				"rect": {
					"x": 0,
					"y": 0,
					"width": 1920,
					"height": 1080
				},
				"output": "HDMI-A-1"
			}
		]`
		var workspaces []sway.Workspace
		err := json.Unmarshal([]byte(workspaceJSON), &workspaces)
		if err != nil {
			t.Fatalf("Failed to parse workspace JSON: %v", err)
		}
		if len(workspaces) != 2 {
			t.Fatalf("Expected 2 workspaces, got %d", len(workspaces))
		}
		ws1 := workspaces[0]
		if ws1.Num != 1 || ws1.Name != "1" {
			t.Errorf("Workspace 1: expected num=1, name='1', got num=%d, name='%s'", ws1.Num, ws1.Name)
		}
		if !ws1.Visible || !ws1.Focused {
			t.Errorf("Workspace 1 should be visible and focused")
		}
		if ws1.Output != "HDMI-A-1" {
			t.Errorf("Workspace 1 output: expected 'HDMI-A-1', got '%s'", ws1.Output)
		}
		ws2 := workspaces[1]
		if ws2.Num != 2 || ws2.Name != "web" {
			t.Errorf("Workspace 2: expected num=2, name='web', got num=%d, name='%s'", ws2.Num, ws2.Name)
		}
		if ws2.Visible || ws2.Focused {
			t.Errorf("Workspace 2 should not be visible or focused")
		}
	})

	t.Run("parse complex sway tree JSON", func(t *testing.T) {
		// test with complex tree structure including idle_inhibitors
		treeJSON := `{
			"id": 1,
			"name": "root",
			"type": "root",
			"layout": "splith",
			"orientation": "horizontal",
			"urgent": false,
			"focused": false,
			"nodes": [
				{
					"id": 2,
					"name": "HDMI-A-1",
					"type": "output",
					"layout": "none",
					"orientation": "none",
					"urgent": false,
					"focused": false,
					"nodes": [
						{
							"id": 3,
							"name": "1",
							"type": "workspace",
							"layout": "splith",
							"orientation": "horizontal",
							"urgent": false,
							"focused": true,
							"nodes": [],
							"floating_nodes": [],
							"idle_inhibitors": {
								"user": "none",
								"application": "none"
							}
						}
					],
					"floating_nodes": []
				}
			],
			"floating_nodes": []
		}`

		var tree sway.Node
		err := json.Unmarshal([]byte(treeJSON), &tree)
		if err != nil {
			t.Fatalf("Failed to parse tree JSON: %v", err)
		}

		// validate root node
		if tree.Type != "root" || tree.Name != "root" {
			t.Errorf("Root node: expected type='root', name='root', got type='%s', name='%s'", tree.Type, tree.Name)
		}

		if len(tree.Nodes) != 1 {
			t.Fatalf("Root should have 1 child node, got %d", len(tree.Nodes))
		}

		// validate output node
		output := tree.Nodes[0]
		if output.Type != "output" || output.Name != "HDMI-A-1" {
			t.Errorf("Output node: expected type='output', name='HDMI-A-1', got type='%s', name='%s'", output.Type, output.Name)
		}

		if len(output.Nodes) != 1 {
			t.Fatalf("Output should have 1 workspace, got %d", len(output.Nodes))
		}

		// validate workspace node
		workspace := output.Nodes[0]
		if workspace.Type != "workspace" || workspace.Name != "1" {
			t.Errorf("Workspace node: expected type='workspace', name='1', got type='%s', name='%s'", workspace.Type, workspace.Name)
		}

		if !workspace.Focused {
			t.Error("Workspace should be focused")
		}

		if workspace.Idle == nil {
			t.Error("idle_inhibitors field should not be nil")
		}

		t.Log("Successfully parsed complex sway tree with idle_inhibitors")
	})
}

// tests that sway commands don't hang indefinitely
func TestSwayCommandTimeout(t *testing.T) {
	if _, err := exec.LookPath("swaymsg"); err != nil {
		t.Skip("swaymsg not available, skipping timeout tests")
	}

	client, err := sway.NewClient()
	if err != nil {
		t.Fatalf("Failed to create sway client: %v", err)
	}

	t.Run("version command completes quickly", func(t *testing.T) {
		start := time.Now()
		_, err := client.GetVersion()
		duration := time.Since(start)

		// allow up to 10 seconds
		maxDuration := 10 * time.Second
		if duration > maxDuration {
			t.Errorf("GetVersion took too long: %v (max: %v)", duration, maxDuration)
		}

		if err != nil {
			t.Logf("GetVersion failed in %v (expected in non-sway environment): %v", duration, err)
		} else {
			t.Logf("GetVersion completed in %v", duration)
		}
	})

	t.Run("tree command completes quickly", func(t *testing.T) {
		start := time.Now()
		_, err := client.GetTree()
		duration := time.Since(start)

		maxDuration := 10 * time.Second
		if duration > maxDuration {
			t.Errorf("GetTree took too long: %v (max: %v)", duration, maxDuration)
		}

		if err != nil {
			t.Logf("GetTree failed in %v (expected in non-sway environment): %v", duration, err)
		} else {
			t.Logf("GetTree completed in %v", duration)
		}
	})
}

// tests graceful error handling
func TestSwayErrorHandling(t *testing.T) {
	t.Run("handles sway not running", func(t *testing.T) {
		if _, err := exec.LookPath("swaymsg"); err != nil {
			t.Skip("swaymsg not available, skipping error handling tests")
		}

		client, err := sway.NewClient()
		if err != nil {
			t.Fatalf("Failed to create sway client: %v", err)
		}

		_, versionErr := client.GetVersion()
		_, treeErr := client.GetTree()
		_, workspacesErr := client.GetWorkspaces()

		// these will fail in non-sway environments

		errors := []error{versionErr, treeErr, workspacesErr}
		methods := []string{"GetVersion", "GetTree", "GetWorkspaces"}

		for i, err := range errors {
			if err != nil {
				// Error should be informative
				errStr := strings.ToLower(err.Error())
				if !strings.Contains(errStr, "sway") &&
				   !strings.Contains(errStr, "connect") &&
				   !strings.Contains(errStr, "ipc") &&
				   !strings.Contains(errStr, "socket") {
					t.Errorf("%s error should be informative: %v", methods[i], err)
				} else {
					t.Logf("%s failed gracefully: %v", methods[i], err)
				}
			} else {
				t.Logf("%s succeeded (running in actual sway)", methods[i])
			}
		}
	})
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func setEnv(key, value string) {
	os.Setenv(key, value)
}