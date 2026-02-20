package layout

import (
	"time"
)

// a saved workspace layout configuration
type Preset struct {
	Name string`json:"name"`
	CreatedAt time.Time `json:"created_at"`
	SwayVersion string `json:"sway_version"`
	Outputs []OutputLayout `json:"outputs"`
	Workspaces []WorkspaceLayout `json:"workspaces"`
}

// the monitor/output configuration
type OutputLayout struct {
	Name string `json:"name"`
	Active bool `json:"active"`
	Rect Rect `json:"rect"`
}

// workspace configuration
type WorkspaceLayout struct {
	Num int `json:"num"`
	Name string `json:"name"`
	Output string `json:"output"`
	Layout string `json:"layout"`
	Containers []ContainerLayout `json:"containers"`
}

// container
type ContainerLayout struct {
	ID int64 `json:"id"`
	Type string `json:"type"`
	Layout string `json:"layout"`
	Orientation string `json:"orientation,omitempty"`
	Percent float64 `json:"percent,omitempty"`
	AppID string `json:"app_id,omitempty"`
	WindowClass string `json:"window_class,omitempty"`
	WindowTitle string `json:"window_title,omitempty"`
	Shell string `json:"shell,omitempty"`
	PID int `json:"pid,omitempty"`
	ExecCommand string `json:"exec_command,omitempty"`
	Floating string `json:"floating,omitempty"`
	Fullscreen int `json:"fullscreen,omitempty"`
	Rect Rect `json:"rect"`
	WindowRect Rect `json:"window_rect"`
	Children []ContainerLayout `json:"children,omitempty"`
	Focused bool `json:"focused"`
}

// rectangle with position and dimensions
type Rect struct {
	X int `json:"x"`
	Y int `json:"y"`
	Width int `json:"width"`
	Height int `json:"height"`
}

// is the same as Preset but summarized for listing purposes
type Metadata struct {
	Name string `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	WorkspaceCount int `json:"workspace_count"`
	WindowCount int `json:"window_count"`
	SwayVersion string `json:"sway_version"`
}
