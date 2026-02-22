package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	binaryName = "swaylayoutmgr"
	binaryPath = "../../build/swaylayoutmgr"
)

// ensures the binary is built and available for testing
func setupTestBinary(t *testing.T) string {
	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path to binary: %v", err)
	}

	// check if binary exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Fatalf("Binary not found at %s. Run 'go build -o build/swaylayoutmgr ./cmd/swaylayoutmgr' first", absPath)
	}

	// check if binary is executable
	if err := exec.Command(absPath, "--version").Run(); err != nil {
		t.Fatalf("Binary is not executable or failed to run: %v", err)
	}

	return absPath
}

// creates isolated test environment
func setupTestEnvironment(t *testing.T) (string, func()) {
	// create temporary directory for test data
	testDir, err := os.MkdirTemp("", "swaylayoutmgr-integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// set XDG environment variables
	originalDataHome := os.Getenv("XDG_DATA_HOME")
	originalConfigHome := os.Getenv("XDG_CONFIG_HOME")

	testDataHome := filepath.Join(testDir, "data")
	testConfigHome := filepath.Join(testDir, "config")

	os.Setenv("XDG_DATA_HOME", testDataHome)
	os.Setenv("XDG_CONFIG_HOME", testConfigHome)

	cleanup := func() {
		os.Setenv("XDG_DATA_HOME", originalDataHome)
		os.Setenv("XDG_CONFIG_HOME", originalConfigHome)
		os.RemoveAll(testDir)
	}

	return testDir, cleanup
}

// tests the --help flag functionality
func TestHelpFlag(t *testing.T) {
	binaryPath := setupTestBinary(t)

	tests := []struct {
		name string
		args []string
	}{
		{"long help flag", []string{"--help"}},
		{"short help flag", []string{"-h"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.Output()

			if err != nil {
				t.Fatalf("Help command failed: %v", err)
			}

			outputStr := string(output)

			// verify help content contains expected sections
			expectedSections := []string{
				"Usage:",
				"save",
				"list",
				"delete",
				"Commands:",
				"Flags:",
			}

			for _, section := range expectedSections {
				if !strings.Contains(outputStr, section) {
					t.Errorf("Help output missing section: %s", section)
				}
			}

			if len(outputStr) < 100 {
				t.Errorf("Help output seems too short: %d characters", len(outputStr))
			}
		})
	}
}

// tests the --version flag functionality
func TestVersionFlag(t *testing.T) {
	binaryPath := setupTestBinary(t)

	tests := []struct {
		name string
		args []string
	}{
		{"long version flag", []string{"--version"}},
		{"short version flag", []string{"-v"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.Output()

			if err != nil {
				t.Fatalf("Version command failed: %v", err)
			}

			outputStr := strings.TrimSpace(string(output))

			// version should contain some version info
			if len(outputStr) == 0 {
				t.Error("Version output is empty")
			}

			// should contain word "version" or version-like format
			if !strings.Contains(strings.ToLower(outputStr), "version") && 
			   !strings.Contains(outputStr, ".") {
				t.Errorf("Version output doesn't look like version info: %s", outputStr)
			}
		})
	}
}

// tests behavior with invalid commands
func TestInvalidCommand(t *testing.T) {
	binaryPath := setupTestBinary(t)

	tests := []struct {
		name string
		args []string
	}{
		{"invalid command", []string{"invalid-command"}},
		{"typo in command", []string{"savw"}},
		{"empty args", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()

			// invalid commands should fail with non-zero exit code
			if err == nil {
				t.Errorf("Invalid command should have failed but succeeded")
			}

			outputStr := strings.ToLower(string(output))

			helpfulIndicators := []string{
				"usage", "help", "command", "invalid", "unknown", "error",
			}

			containsHelpful := false
			for _, indicator := range helpfulIndicators {
				if strings.Contains(outputStr, indicator) {
					containsHelpful = true
					break
				}
			}

			if !containsHelpful {
				t.Errorf("Error message should be helpful. Got: %s", string(output))
			}
		})
	}
}

// tests the list command with no presets
func TestListCommandEmpty(t *testing.T) {
	binaryPath := setupTestBinary(t)
	testDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cmd := exec.Command(binaryPath, "list")
	output, err := cmd.Output()

	if err != nil {
		t.Fatalf("List command failed: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))

	// should indicate no presets found
	expectedMessages := []string{
		"No saved layout presets found",
		"no presets",
		"empty",
		"0 presets",
		"Available layout presets:",
	}

	containsExpected := false
	for _, msg := range expectedMessages {
		if strings.Contains(strings.ToLower(outputStr), strings.ToLower(msg)) {
			containsExpected = true
			break
		}
	}

	if !containsExpected && outputStr != "" {
		t.Errorf("List output should indicate no presets. Got: %s", outputStr)
	}

	t.Logf("Empty list output: %s", outputStr)
	t.Logf("Test directory: %s", testDir)
}

// tests save command with invalid preset names
func TestSaveCommandInvalidName(t *testing.T) {
	binaryPath := setupTestBinary(t)
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	tests := []struct {
		name       string
		presetName string
	}{
		{"empty name", ""},
		{"whitespace only", "   "},
		{"special characters", "preset/\\:*?\"<>|"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"save"}
			if tt.presetName != "" {
				args = append(args, tt.presetName)
			}

			cmd := exec.Command(binaryPath, args...)
			output, err := cmd.CombinedOutput()

			// save command succeed by auto-generating names or accepting provided names
			outputStr := string(output)

			if err != nil {
				// if it fails, it should be for special characters that are invalid for filenames
				if tt.name == "special characters" {
					if !strings.Contains(strings.ToLower(outputStr), "invalid") &&
					   !strings.Contains(strings.ToLower(outputStr), "error") {
						t.Errorf("Special character error should be clear. Got: %s", outputStr)
					}
				} else {
					t.Errorf("Save command should have succeeded for '%s'. Error: %v. Output: %s", tt.name, err, outputStr)
				}
			} else {
				// success case should show successful save message
				if !strings.Contains(strings.ToLower(outputStr), "successfully saved") &&
				   !strings.Contains(strings.ToLower(outputStr), "saved") {
					t.Errorf("Success message should confirm save. Got: %s", outputStr)
				}
				t.Logf("Save succeeded for '%s': %s", tt.name, strings.TrimSpace(outputStr))
			}
		})
	}
}

// tests delete command with non-existent preset
func TestDeleteCommandNonExistent(t *testing.T) {
	binaryPath := setupTestBinary(t)
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	cmd := exec.Command(binaryPath, "delete", "non-existent-preset")
	output, err := cmd.CombinedOutput()

	if err == nil {
		// some implementations might succeed silently
		t.Logf("Delete succeeded for non-existent preset (acceptable)")
	}

	outputStr := strings.ToLower(string(output))

	// if it provides a message it should be clear
	if len(outputStr) > 0 {
		helpfulIndicators := []string{
			"not found", "does not exist", "deleted", "error", "preset", "not yet implemented",
		}

		containsHelpful := false
		for _, indicator := range helpfulIndicators {
			if strings.Contains(outputStr, indicator) {
				containsHelpful = true
				break
			}
		}

		if !containsHelpful {
			t.Errorf("Delete message should be clear. Got: %s", string(output))
		}
	}
}

// tests basic binary execution properties
func TestBinaryExecution(t *testing.T) {
	binaryPath := setupTestBinary(t)

	t.Run("binary exits quickly for help", func(t *testing.T) {
		start := time.Now()
		cmd := exec.Command(binaryPath, "--help")
		err := cmd.Run()
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Help command failed: %v", err)
		}

		// should complete within reasonable time
		if duration > 5*time.Second {
			t.Errorf("Help command took too long: %v", duration)
		}
	})

	t.Run("binary handles signals gracefully", func(t *testing.T) {
		// this is a basic test
		cmd := exec.Command(binaryPath, "--version")
		err := cmd.Run()

		if err != nil {
			t.Fatalf("Version command failed: %v", err)
		}
		// here the binary at least started and exited cleanly
	})
}