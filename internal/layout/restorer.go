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
	forWindowRules []string // track for_window rules to remove after restoration
}

// contains the outcome of a restoration operation
type RestoreResult struct {
	WorkspacesRestored int
	ApplicationsLaunched int
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
	r.forWindowRules = []string{} // reset rules list
	result := &RestoreResult{}

	// restore each workspace without switching focus
	for _, workspace := range preset.Workspaces {
		if err := r.restoreWorkspaceInBackground(&workspace, result); err != nil {
			// collect error but continue with other workspaces
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

	// remove all for_window rules after restoration is complete
	r.removeForWindowRules()

	result.Errors = r.errors
	return result, nil
}

// restores a workspace in the background without switching focus
func (r *Restorer) restoreWorkspaceInBackground(workspace *WorkspaceLayout, result *RestoreResult) error {
	for i := range workspace.Containers {
		if err := r.restoreContainerInBackground(&workspace.Containers[i], workspace, result); err != nil {
			// collect error but continue with other containers
			containerID := workspace.Containers[i].AppID
			if containerID == "" {
				containerID = workspace.Containers[i].WindowClass
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

// recursively restores containers in the background without focus changes
func (r *Restorer) restoreContainerInBackground(container *ContainerLayout, workspace *WorkspaceLayout, result *RestoreResult) error {
	// if container has an executable command
	if container.ExecCommand != "" {
		return r.launchApplicationInBackground(container, workspace, result)
	}

	// if this is a parent container with children, restore them recursively
	if len(container.Children) > 0 {
		for i := range container.Children {
			if err := r.restoreContainerInBackground(&container.Children[i], workspace, result); err != nil {
				// collect error but continue with siblings
				r.errors = append(r.errors, err)
			}
		}
	}
	return nil
}

// launches an application in the background and moves it to the target workspace
// This mimics Sway's assign command - windows are moved without changing focus
func (r *Restorer) launchApplicationInBackground(container *ContainerLayout, workspace *WorkspaceLayout, result *RestoreResult) error {
	if container.ExecCommand == "" {
		return fmt.Errorf("no exec command available for app_id=%s class=%s", container.AppID, container.WindowClass)
	}

	// check if application is already running (only if ReuseExistingWindows is enabled)
	if r.ReuseExistingWindows && r.isApplicationRunning(container) {
		// move existing window to correct workspace (doesn't change focus)
		if err := r.moveWindowToWorkspace(container, workspace); err != nil {
			windowID := container.AppID
			if windowID == "" {
				windowID = container.WindowClass
			}
			r.errors = append(r.errors, fmt.Errorf("failed to move existing window '%s': %w", windowID, err))
		}
		result.ApplicationsLaunched++
		return nil
	}

	criteria := r.buildWindowCriteria(container)

	// Build the target workspace identifier
	var workspaceTarget string
	if workspace.Num < 0 {
		workspaceTarget = workspace.Name
	} else {
		workspaceTarget = fmt.Sprintf("number %d", workspace.Num)
	}

	// Set up for_window rule to auto-move window to target workspace when it spawns
	// This prevents the window from appearing on the current workspace
	forWindowCmd := fmt.Sprintf("for_window [%s] move container to workspace %s", criteria, workspaceTarget)
	if err := r.client.RunCommand(forWindowCmd); err != nil {
		return fmt.Errorf("failed to set for_window rule: %w", err)
	}

	// Track this rule so we can remove it later
	r.forWindowRules = append(r.forWindowRules, criteria)

	// launch the application on the temp workspace
	execCmd := fmt.Sprintf("exec %s", container.ExecCommand)
	if err := r.client.RunCommand(execCmd); err != nil {
		return fmt.Errorf("failed to launch %s: %w", container.ExecCommand, err)
	}

	// wait for window to appear (it will be auto-moved by the for_window rule)
	if err := r.waitForWindow(container, 10*time.Second); err != nil {
		return fmt.Errorf("window did not appear: %w", err)
	}

	// restore window properties (floating, position, size)
	if err := r.restoreWindowProperties(container); err != nil {
		windowID := container.AppID
		if windowID == "" {
			windowID = container.WindowClass
		}
		r.errors = append(r.errors, fmt.Errorf("failed to restore properties for '%s': %w", windowID, err))
	}

	result.ApplicationsLaunched++
	return nil
}

// checks if an application is already running
func (r *Restorer) isApplicationRunning(container *ContainerLayout) bool {
	tree, err := r.client.GetTree()
	if err != nil {
		return false
	}

	return r.findWindowInTree(tree, container) != nil
}

// finds a window in the tree matching the container
func (r *Restorer) findWindowInTree(node *sway.Node, container *ContainerLayout) *sway.Node {
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

// moves an existing window to the specified workspace
func (r *Restorer) moveWindowToWorkspace(container *ContainerLayout, workspace *WorkspaceLayout) error {
	criteria := r.buildWindowCriteria(container)
	var moveCmd string
	if workspace.Num < 0 {
		// named workspace
		moveCmd = fmt.Sprintf("[%s] move to workspace %s", criteria, workspace.Name)
	} else {
		// numbered workspace
		moveCmd = fmt.Sprintf("[%s] move to workspace number %d", criteria, workspace.Num)
	}
	return r.client.RunCommand(moveCmd)
}

// restores window properties
func (r *Restorer) restoreWindowProperties(container *ContainerLayout) error {
	criteria := r.buildWindowCriteria(container)

	// restore floating state
	if container.Floating != "" && container.Floating != "auto_off" {
		floatingCmd := fmt.Sprintf("[%s] floating enable", criteria)
		if err := r.client.RunCommand(floatingCmd); err != nil {
			return fmt.Errorf("failed to set floating: %w", err)
		}

		// for floating windows, restore position and size
		if container.Rect.Width > 0 && container.Rect.Height > 0 {
			windowID := container.AppID
			if windowID == "" {
				windowID = container.WindowClass
			}
			moveCmd := fmt.Sprintf("[%s] move position %d %d", criteria, container.Rect.X, container.Rect.Y)
			if err := r.client.RunCommand(moveCmd); err != nil {
				r.errors = append(r.errors, fmt.Errorf("failed to move window '%s': %w", windowID, err))
			}

			resizeCmd := fmt.Sprintf("[%s] resize set %d %d", criteria, container.Rect.Width, container.Rect.Height)
			if err := r.client.RunCommand(resizeCmd); err != nil {
				r.errors = append(r.errors, fmt.Errorf("failed to resize window '%s': %w", windowID, err))
			}
		}
	}

	return nil
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

// removes all for_window rules that were set during restoration
func (r *Restorer) removeForWindowRules() {
	if err := r.client.RunCommand("reload"); err != nil {
		// Non-fatal - just log the error
		r.errors = append(r.errors, fmt.Errorf("failed to reload config to clear for_window rules: %w", err))
	}
}