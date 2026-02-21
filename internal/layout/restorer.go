package layout

import (
	"fmt"
	"strings"
	"time"
	"github.com/lidsol/sway-layout-manager/internal/sway"
)

// handles restoring layouts in Sway
type Restorer struct {
	client *sway.Client
	errors []error // non-fatal errors during restoration
	ReuseExistingWindows bool
	ReuseApps []string
	reusedWindowIDs map[int64]bool // tracks IDs of windows that have been reused
}

// contains the outcome of a restoration operation
type RestoreResult struct {
	WorkspacesRestored int
	ApplicationsLaunched int
	ApplicationsReused int
	ApplicationsFailed int
	Errors []error
}

// constructor
func NewRestorer(client *sway.Client) *Restorer {
	return &Restorer{
		client: client,
	}
}

// applies a saved preset to the current Sway session
func (r *Restorer) Restore(preset *Preset) (*RestoreResult, error) {
	if preset == nil {
		return nil, fmt.Errorf("preset cannot be nil")
	}

	r.errors = []error{}
	r.reusedWindowIDs = make(map[int64]bool)
	result := &RestoreResult{}

	// save current workspace to return to it after restoration
	workspaces, err := r.client.GetWorkspaces()
	if err == nil {
		for _, ws := range workspaces {
			if ws.Focused {
				defer r.client.RunCommand(fmt.Sprintf("workspace number %d", ws.Num))
				break
			}
		}
	}

	// restore each workspace
	for _, workspace := range preset.Workspaces {
		if err := r.restoreWorkspace(&workspace, result); err != nil {
			if workspace.Num < 0 {
				r.errors = append(r.errors, fmt.Errorf("workspace '%s': %w", workspace.Name, err))
			} else {
				r.errors = append(r.errors, fmt.Errorf("workspace %d: %w", workspace.Num, err))
			}
		} else {
			result.WorkspacesRestored++
		}
	}

	// verify restoration
	if err := r.verifyRestoration(preset); err != nil {
		r.errors = append(r.errors, fmt.Errorf("verification: %w", err))
	}

	result.Errors = r.errors
	return result, nil
}

// restores a workspace by switching to it and arranging windows directly
func (r *Restorer) restoreWorkspace(workspace *WorkspaceLayout, result *RestoreResult) error {
	if len(workspace.Containers) == 0 {
		return nil
	}

	// switch to the target workspace
	workspaceID := r.getWorkspaceIdentifier(workspace)
	if err := r.client.RunCommand(fmt.Sprintf("workspace %s", workspaceID)); err != nil {
		return fmt.Errorf("failed to switch to workspace: %w", err)
	}
	time.Sleep(100 * time.Millisecond)

	// restore each top-level container
	for i := range workspace.Containers {
		container := &workspace.Containers[i]

		// if not the first container, prepare for a sibling split
		if i > 0 {
			// wait for previous containers
			prevContainer := &workspace.Containers[i-1]
			r.waitForContainerWindows(prevContainer, 5*time.Second)

			// delay to ensure window appearance
			time.Sleep(300 * time.Millisecond)

			// refocus the workspace to ensure
			r.client.RunCommand(fmt.Sprintf("workspace %s", workspaceID))
			time.Sleep(100 * time.Millisecond)

			// ensures the next split happens at the workspace level
			r.client.RunCommand("focus parent")
			time.Sleep(50 * time.Millisecond)

			// split
			splitDir := "h"
			if workspace.Layout == "splitv" {
				splitDir = "v"
			}
			r.client.RunCommand(fmt.Sprintf("split %s", splitDir))
			time.Sleep(100 * time.Millisecond)
		}

		// restore this container
		if err := r.restoreContainerDirect(container, result); err != nil {
			containerID := container.AppID
			if containerID == "" {
				containerID = container.WindowClass
			}
			if containerID != "" {
				r.errors = append(r.errors, fmt.Errorf("container '%s': %w", containerID, err))
			} else {
				r.errors = append(r.errors, err)
			}
			result.ApplicationsFailed++
		}
	}

	return nil
}

// waits for all windows in a container to appear
func (r *Restorer) waitForContainerWindows(container *ContainerLayout, timeout time.Duration) {
	// collect all windows that should appear
	var windows []*ContainerLayout
	r.collectLeafWindows(container, &windows)

	if len(windows) == 0 {
		return
	}

	deadline := time.Now().Add(timeout)

	// wait for each window to appear
	for _, win := range windows {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}

		// wait for this specific window with remaining timeout
		if err := r.waitForWindow(win, remaining); err != nil {
			if windowID := r.getContainerID(win); windowID != "" {
				r.errors = append(r.errors, fmt.Errorf("window '%s' may not have appeared: %w", windowID, err))
			}
		}
	}
}

// restores a container directly on the currently focused workspace
func (r *Restorer) restoreContainerDirect(container *ContainerLayout, result *RestoreResult) error {
	// if this container is a single window, launch it
	if container.ExecCommand != "" {
		return r.launchApplicationDirect(container, result)
	}

	// if this is a grouping container (tabbed/stacked), restore specially
	if container.Layout == "tabbed" || container.Layout == "stacked" {
		return r.restoreGroupedContainerDirect(container, result)
	}

	// for split containers, restore children
	for i := range container.Children {
		if i > 0 {
			// create split between siblings
			splitDir := "h"
			if container.Layout == "splitv" {
				splitDir = "v"
			}
			r.client.RunCommand(fmt.Sprintf("split %s", splitDir))
		}

		if err := r.restoreContainerDirect(&container.Children[i], result); err != nil {
			r.errors = append(r.errors, err)
		}
	}

	return nil
}

// restores a grouped container directly
func (r *Restorer) restoreGroupedContainerDirect(container *ContainerLayout, result *RestoreResult) error {
	var leafWindows []*ContainerLayout
	r.collectLeafWindows(container, &leafWindows)

	if len(leafWindows) == 0 {
		return nil
	}

	// launch first window
	if err := r.launchApplicationDirect(leafWindows[0], result); err != nil {
		return err
	}

	// set layout to tabbed/stacked on the current container
	if err := r.client.RunCommand(fmt.Sprintf("layout %s", container.Layout)); err != nil {
		return fmt.Errorf("failed to set layout: %w", err)
	}

	// launch remaining windows (they'll automatically join the tabbed/stacked container)
	for i := 1; i < len(leafWindows); i++ {
		if err := r.launchApplicationDirect(leafWindows[i], result); err != nil {
			r.errors = append(r.errors, err)
		}
	}

	return nil
}

// launches an application directly on the current workspace/container
func (r *Restorer) launchApplicationDirect(container *ContainerLayout, result *RestoreResult) error {
	if container.ExecCommand == "" {
		return fmt.Errorf("no exec command available")
	}

	// check if application is already running and reuse if enabled
	if r.ReuseExistingWindows || len(r.ReuseApps) > 0 {
		if window := r.findAvailableWindow(container); window != nil {
			// set the app as reused
			r.reusedWindowIDs[window.ID] = true
			result.ApplicationsReused++
			return nil
		}
	}

	// launch the application
	if err := r.client.RunCommand(fmt.Sprintf("exec %s", container.ExecCommand)); err != nil {
		return fmt.Errorf("failed to launch %s: %w", container.ExecCommand, err)
	}

	// brief delay to allow window to start
	time.Sleep(400 * time.Millisecond)

	if r.ReuseExistingWindows || len(r.ReuseApps) > 0 {
		if window := r.findAvailableWindow(container); window != nil {
			r.reusedWindowIDs[window.ID] = true
		}
	}

	result.ApplicationsLaunched++
	return nil
}

// collects all leaf windows (with exec commands) from a container tree
func (r *Restorer) collectLeafWindows(container *ContainerLayout, result *[]*ContainerLayout) {
	if container.ExecCommand != "" {
		*result = append(*result, container)
		return
	}

	for i := range container.Children {
		r.collectLeafWindows(&container.Children[i], result)
	}
}

// finds an available (not yet reused) window for the given container
func (r *Restorer) findAvailableWindow(container *ContainerLayout) *sway.Node {
	tree, err := r.client.GetTree()
	if err != nil {
		return nil
	}

	return r.findWindowInTree(tree, container)
}

// checks if an application is running (regardless of reuse status)
func (r *Restorer) isApplicationRunning(container *ContainerLayout) bool {
	tree, err := r.client.GetTree()
	if err != nil {
		return false
	}

	return r.findAnyWindowInTree(tree, container) != nil
}

// finds a window in the tree matching the container
func (r *Restorer) findWindowInTree(node *sway.Node, container *ContainerLayout) *sway.Node {
	// extract the window class if available
	var windowClass string
	if node.WindowProperties != nil {
		windowClass = node.WindowProperties.Class
	}

	// check if this node matches the container's app_id or class
	nodeMatches := false
	if container.AppID != "" && node.AppID == container.AppID {
		nodeMatches = true
	}
	if container.WindowClass != "" && windowClass == container.WindowClass {
		nodeMatches = true
	}

	// if the node matches and should be reused, return it
	if nodeMatches && !r.reusedWindowIDs[node.ID] && r.shouldReuseWindow(node.AppID, windowClass) {
		return node
	}

	// search children
	for i := range node.Nodes {
		if found := r.findWindowInTree(&node.Nodes[i], container); found != nil {
			return found
		}
	}
	for i := range node.FloatingNodes {
		if found := r.findWindowInTree(&node.FloatingNodes[i], container); found != nil {
			return found
		}
	}

	return nil
}

// determines if a window should be reused
func (r *Restorer) shouldReuseWindow(appID string, windowClass string) bool {
	// if --reuse-all is enabled
	if r.ReuseExistingWindows {
		return true
	}

	// if --reuse is empty
	if len(r.ReuseApps) == 0 {
		return false
	}

	// check if the window matches any of the ReuseApps criteria
	for _, criteria := range r.ReuseApps {
		if criteria == appID || criteria == windowClass {
			return true
		}
	}

	return false
}

// finds any window in the tree matching the container
func (r *Restorer) findAnyWindowInTree(node *sway.Node, container *ContainerLayout) *sway.Node {
	// check if this node matches
	if container.AppID != "" && node.AppID == container.AppID {
		return node
	}
	if container.WindowClass != "" && node.WindowProperties != nil &&
		node.WindowProperties.Class == container.WindowClass {
		return node
	}

	// search children
	for i := range node.Nodes {
		if found := r.findAnyWindowInTree(&node.Nodes[i], container); found != nil {
			return found
		}
	}
	for i := range node.FloatingNodes {
		if found := r.findAnyWindowInTree(&node.FloatingNodes[i], container); found != nil {
			return found
		}
	}

	return nil
}

// waits for a window to appear in the tree
func (r *Restorer) waitForWindow(container *ContainerLayout, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if r.isApplicationRunning(container) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for window %s/%s", container.AppID, container.WindowClass)
}

// builds a window selection criteria string
func (r *Restorer) buildWindowCriteria(container *ContainerLayout) string {
	var criteria []string

	if container.AppID != "" {
		criteria = append(criteria, fmt.Sprintf("app_id=\"%s\"", container.AppID))
	}
	if container.WindowClass != "" {
		criteria = append(criteria, fmt.Sprintf("class=\"%s\"", container.WindowClass))
	}

	return strings.Join(criteria, " ")
}

// verifies that the restoration was successful
func (r *Restorer) verifyRestoration(preset *Preset) error {
	// collect all containers with exec commands
	var expectedContainers []*ContainerLayout
	for _, ws := range preset.Workspaces {
		r.collectContainersWithExec(&ws.Containers, &expectedContainers)
	}

	// check which containers are missing
	var missingContainers []string
	for _, container := range expectedContainers {
		if !r.isApplicationRunning(container) {
			windowID := container.AppID
			if windowID == "" {
				windowID = container.WindowClass
			}
			if windowID != "" {
				missingContainers = append(missingContainers, windowID)
			}
		}
	}

	if len(missingContainers) > 0 {
		return fmt.Errorf("missing windows: %s", strings.Join(missingContainers, ", "))
	}

	return nil
}

// collects containers with exec commands recursively
func (r *Restorer) collectContainersWithExec(containers *[]ContainerLayout, result *[]*ContainerLayout) {
	for i := range *containers {
		if (*containers)[i].ExecCommand != "" {
			*result = append(*result, &(*containers)[i])
		}
		if len((*containers)[i].Children) > 0 {
			r.collectContainersWithExec(&(*containers)[i].Children, result)
		}
	}
}

// helper function to get workspace identifie)
func (r *Restorer) getWorkspaceIdentifier(workspace *WorkspaceLayout) string {
	if workspace.Num < 0 {
		return workspace.Name
	}
	return fmt.Sprintf("number %d", workspace.Num)
}

// helper function to get container ID
func (r *Restorer) getContainerID(container *ContainerLayout) string {
	if container.AppID != "" {
		return container.AppID
	}
	return container.WindowClass
}