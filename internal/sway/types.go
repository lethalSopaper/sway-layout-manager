package sway

// node in the sway tree
type Node struct {
	ID int64 `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Layout string `json:"layout"`
	Orientation string  `json:"orientation"`
	Percent float64 `json:"percent,omitempty"`
	Urgent bool `json:"urgent"`
	Focused bool `json:"focused"`
	Output string `json:"output,omitempty"`
	Floating string `json:"floating,omitempty"`
	Fullscreen int `json:"fullscreen_mode"`
	AppID string `json:"app_id,omitempty"`
	PID int `json:"pid,omitempty"`
	Visible bool `json:"visible"`
	Shell string `json:"shell,omitempty"`
	Inhibit bool `json:"inhibit_idle"`
	Idle bool `json:"idle_inhibitors"`
	Window int `json:"window,omitempty"`
	WindowProperties *WindowProps `json:"window_properties,omitempty"`
	Nodes []Node `json:"nodes,omitempty"`
	FloatingNodes []Node `json:"floating_nodes,omitempty"`
	Rect Rect `json:"rect"`
	WindowRect Rect `json:"window_rect"`
	DecoRect Rect `json:"deco_rect"`
	Geometry Rect `json:"geometry"`
	Focus []int64 `json:"focus"`
	Sticky bool `json:"sticky"`
	Num int `json:"num,omitempty"`
}

// X11 window properties
type WindowProps struct {
	Class string `json:"class,omitempty"`
	Instance string `json:"instance,omitempty"`
	Title string `json:"title,omitempty"`
	TransientFor int `json:"transient_for,omitempty"`
}

// rectangle with position and size
type Rect struct {
	X int `json:"x"`
	Y int `json:"y"`
	Width int `json:"width"`
	Height int `json:"height"`
}

// sway workspace
type Workspace struct {
	ID int64 `json:"id"`
	Num int `json:"num"`
	Name string `json:"name"`
	Visible bool `json:"visible"`
	Focused bool `json:"focused"`
	Urgent bool `json:"urgent"`
	Rect Rect `json:"rect"`
	Output string `json:"output"`
}

// sway version information
type Version struct {
	HumanReadable string `json:"human_readable"`
	Variant string `json:"variant"`
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
	LoadedConfigFileName string `json:"loaded_config_file_name"`
}
