# Sway Layout Manager

**Workspace layout manager for Sway WM**

Sway Layout Manager is an application that automates the process of saving and restoring window layouts and container arrangements in Sway. Save your current workspace arrangements as named presets and restore them when needed, eliminating the tedious task of manually reopening and positioning applications.

## Features

- **Save Layout Presets**: Get current window arrangements across multiple workspaces and outputs
- **Layout Loading**: Load saved layouts restoring window positions and applications
- **Dual Interface**: Command-line interface (CLI) for automation and scripting
- **Future GUI Support**: Planned graphical interface for visual management
- **No Dependencies**: Core CLI compiles to a single executable with no runtime dependencies

## Requirements

- **Operating System**: GNU/Linux distribution with Wayland support
- **Window Manager**: Sway (version 1.10 or higher)
- **Runtime**: No additional dependencies for CLI usage

## Installation

> **Note**: This project is currently in development. Installation instructions will be provided once the first release is available.

## Building from Source

### Prerequisites

- Go 1.21 or higher
- Git

### Build Steps

1. Clone the repository:

```bash
git clone https://github.com/lidsol/sway-layout-manager.git
cd sway-layout-manager
```

2. Build the executable:

```bash
go build -o swaylayoutmgr ./cmd/swaylayoutmgr
```

3. (Optional) Install to your system:

```bash
sudo mv swaylayoutmgr /usr/local/bin/
```

## Usage

### Save Current Layout

```bash
swaylayoutmgr save <name>
```

Captures the current state of all active workspaces and saves it as a named layout preset.

### Load Saved Layout

```bash
swaylayoutmgr load <name>
```

Restores a previously saved layout, launching applications and positioning them according to the saved configuration.

### List Available Presets

```bash
swaylayoutmgr list
```

Displays all available layout presets with metadata including creation date, workspace count, and application summary.

### Delete Preset

```bash
swaylayoutmgr delete <name>
```

Removes a saved preset (with confirmation prompt).

## Use Case Example

A developer working with Visual Studio Code on the left half of the screen and two terminal windows stacked vertically on the right half can save this arrangement as `dev-workspace`. Later, they can instantly recreate this exact layout with:

```bash
swaylayoutmgr load dev-workspace
```

## Contributing

> **Note**: Detailed contribution guidelines will be added as the project develops.

## License

> **Note**: License information will be added in future releases.

## Project Status

**In Development** - This project is currently in the proposal and early development phase.
