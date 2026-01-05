package layout

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"
	"github.com/lidsol/sway-layout-manager/internal/sway"
)

// struct to parse the current Sway layout
type Parser struct {
	swayClient *sway.Client
}

// parser contructor
func NewParser(client *sway.Client) *Parser {
	return &Parser{
		swayClient: client,
	}
}

// reads the command line used to launch a process from /proc
func getProcessCommand(pid int) string {
	if pid <= 0 {
		return ""
	}

	// check if this is a Flatpak app
	flatpakID := detectFlatpakApp(pid);
	if flatpakID != "" {
		return fmt.Sprintf("flatpak run %s", flatpakID)
	}

	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
	data, err := os.ReadFile(cmdlinePath)
	if err != nil {
		// process might have exited or no permission
		return ""
	}

	// replace null bytes with spaces for a readable command
	cmdline := string(bytes.ReplaceAll(data, []byte{0}, []byte{' '}))
	cmdline = strings.TrimSpace(cmdline)

	return cmdline
}

// if a process is a Flatpak app, returns its app ID
func detectFlatpakApp(pid int) string {
	environPath := fmt.Sprintf("/proc/%d/environ", pid)
	data, err := os.ReadFile(environPath)
	if err != nil {
		return ""
	}

	// environ uses null bytes as separators
	envVars := bytes.Split(data, []byte{0})
	for _, envVar := range envVars {
		if bytes.HasPrefix(envVar, []byte("FLATPAK_ID=")) {
			appID := string(bytes.TrimPrefix(envVar, []byte("FLATPAK_ID=")))
			return appID
		}
	}

	// fallback: check root directory for .flatpak-info
	rootPath := fmt.Sprintf("/proc/%d/root/.flatpak-info", pid)
	if _, err := os.Stat(rootPath); err == nil {
		infoData, err := os.ReadFile(rootPath)
		if err == nil {
			// parse for Application section with name=
			lines := strings.Split(string(infoData), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "name=") {
					appID := strings.TrimPrefix(line, "name=")
					appID = strings.TrimSpace(appID)
					return appID
				}
			}
		}
	}

	return ""
}

// captures the current Sway layout
func (p *Parser) CaptureCurrentLayout(name string) (*Preset, error) {
	// gets version, tree, and workspaces
	version, err := p.swayClient.GetVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get sway version: %w", err)
	}

	tree, err := p.swayClient.GetTree()
	if err != nil {
		return nil, fmt.Errorf("failed to get window tree: %w", err)
	}

	workspaces, err := p.swayClient.GetWorkspaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces: %w", err)
	}

	preset := &Preset{
		Name: name,
		CreatedAt: time.Now(),
		SwayVersion: version.HumanReadable,
		Outputs: make([]OutputLayout, 0),
		Workspaces: make([]WorkspaceLayout, 0),
	}

	// parse outputs and workspaces
	p.parseOutputs(tree, preset)
	p.parseWorkspaces(tree, workspaces, preset)

	return preset, nil
}

// output information from the tree
func (p *Parser) parseOutputs(node *sway.Node, preset *Preset) {
	// only consider output nodes
	if node.Type == "output" && node.Name != "__i3" {
		output := OutputLayout{
			Name:   node.Name,
			Active: node.Name != "",
			Rect: Rect{
				X: node.Rect.X,
				Y: node.Rect.Y,
				Width: node.Rect.Width,
				Height: node.Rect.Height,
			},
		}

		preset.Outputs = append(preset.Outputs, output)
	}

	// process child nodes
	for _, child := range node.Nodes {
		p.parseOutputs(&child, preset)
	}
}

// workspace and container information
func (p *Parser) parseWorkspaces(tree *sway.Node, workspaces []sway.Workspace, preset *Preset) {
	workspaceNodes := p.findWorkspaceNodes(tree)
	for _, wsNode := range workspaceNodes {
		// finds corresponding workspace info
		var wsInfo *sway.Workspace
		for i := range workspaces {
			if workspaces[i].Num == wsNode.Num {
				wsInfo = &workspaces[i]
				break
			}
		}

		if wsInfo == nil {
			continue
		}

		workspace := WorkspaceLayout{
			Num: wsInfo.Num,
			Name: wsInfo.Name,
			Output: wsInfo.Output,
			Layout: wsNode.Layout,
			Containers: make([]ContainerLayout, 0),
		}

		p.parseContainers(wsNode, &workspace.Containers)
		preset.Workspaces = append(preset.Workspaces, workspace)
	}
}

// findWorkspaceNodes recursively finds all workspace nodes
func (p *Parser) findWorkspaceNodes(node *sway.Node) []*sway.Node {
	var workspaces []*sway.Node

	if node.Type == "workspace" && node.Name != "__i3_scratch" {
		workspaces = append(workspaces, node)
	}

	for i := range node.Nodes {
		workspaces = append(workspaces, p.findWorkspaceNodes(&node.Nodes[i])...)
	}

	return workspaces
}

// parses container hierarchy
func (p *Parser) parseContainers(node *sway.Node, containers *[]ContainerLayout) {
	if node.Type == "workspace" && node.Name == "__i3_scratch" {
		return
	}

	// only process containers and windows
	if node.Type == "con" || node.Type == "floating_con" {
		container := ContainerLayout{
			ID: node.ID,
			Type: node.Type,
			Layout: node.Layout,
			Orientation: node.Orientation,
			Percent: node.Percent,
			AppID: node.AppID,
			Shell: node.Shell,
			PID: node.PID,
			ExecCommand: getProcessCommand(node.PID),
			Floating: node.Floating,
			Fullscreen: node.Fullscreen,
			Focused: node.Focused,
			Rect: Rect{
				X: node.Rect.X,
				Y: node.Rect.Y,
				Width: node.Rect.Width,
				Height: node.Rect.Height,
			},
			WindowRect: Rect{
				X: node.WindowRect.X,
				Y: node.WindowRect.Y,
				Width: node.WindowRect.Width,
				Height: node.WindowRect.Height,
			},
			Children: make([]ContainerLayout, 0),
		}

		// X11 windows
		if node.WindowProperties != nil {
			container.WindowClass = node.WindowProperties.Class
			container.WindowTitle = node.WindowProperties.Title
		}

		if container.WindowTitle == "" && node.Name != "" {
			container.WindowTitle = node.Name
		}

		// parse child containers recursively
		for i := range node.Nodes {
			p.parseContainers(&node.Nodes[i], &container.Children)
		}

		// parse floating windows
		for i := range node.FloatingNodes {
			p.parseContainers(&node.FloatingNodes[i], &container.Children)
		}

		*containers = append(*containers, container)
	} else {
		// for non-container nodes, recurse into children
		for i := range node.Nodes {
			p.parseContainers(&node.Nodes[i], containers)
		}
		for i := range node.FloatingNodes {
			p.parseContainers(&node.FloatingNodes[i], containers)
		}
	}
}
