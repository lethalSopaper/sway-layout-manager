package sway

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// handles communication via swaymsg
type Client struct {
	swayMsgPath string
}

// constructor
func NewClient() (*Client, error) {
	// verifies if swaymsg is available in PATH
	path, err := exec.LookPath("swaymsg")
	if err != nil {
		return nil, fmt.Errorf("swaymsg not found in PATH: %w", err)
	}

	return &Client{
		swayMsgPath: path,
	}, nil
}

// gets the complete window tree from sway in JSON format
func (c *Client) GetTree() (*Node, error) {
	cmd := exec.Command(c.swayMsgPath, "-t", "get_tree", "--raw")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute swaymsg: %w", err)
	}

	var tree Node
	if err := json.Unmarshal(output, &tree); err != nil {
		return nil, fmt.Errorf("failed to parse sway tree: %w", err)
	}

	return &tree, nil
}

// retrieves information about all workspaces
func (c *Client) GetWorkspaces() ([]Workspace, error) {
	cmd := exec.Command(c.swayMsgPath, "-t", "get_workspaces", "--raw")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute swaymsg: %w", err)
	}

	var workspaces []Workspace
	if err := json.Unmarshal(output, &workspaces); err != nil {
		return nil, fmt.Errorf("failed to parse workspaces: %w", err)
	}

	return workspaces, nil
}

// sway version information
func (c *Client) GetVersion() (*Version, error) {
	cmd := exec.Command(c.swayMsgPath, "-t", "get_version", "--raw")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute swaymsg: %w", err)
	}

	var version Version
	if err := json.Unmarshal(output, &version); err != nil {
		return nil, fmt.Errorf("failed to parse version: %w", err)
	}

	return &version, nil
}

// executes a Sway IPC command
func (c *Client) RunCommand(command string) error {
	cmd := exec.Command(c.swayMsgPath, command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sway command failed: %s: %w", string(output), err)
	}
	return nil
}
