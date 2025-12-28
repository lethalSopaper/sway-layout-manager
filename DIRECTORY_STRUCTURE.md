# Directory Structure Documentation

## Project Root Structure

### `/cmd/`

Contains the main applications for this project

- **`swaylayoutmgr/`**: Main CLI application entry point
  - Expected files: `main.go`, command-line interface logic
  - Contains the executable's main function and CLI command definitions

### `/docs/`

Project documentation including proposals, specifications, and guides

- **`proposal/`**: Project proposal and technical specifications
  - Contains documents explaining the project concept

- **`workflow/`**: Development workflow and process documentation
  - Phase-based development documentation

### `/gui/`

Graphical User Interface components

- **`assets/`**: Static assets for GUI (icons, images, stylesheets)

- **`components/`**: Reusable GUI components and widgets

- **`main/`**: Main GUI application entry point and window management

### `/internal/`

Private application code that shouldn't be imported by other projects

- **`config/`**: Configuration management and settings
  - Config file parsing, validation, and default values
  - User preferences and application settings

- **`layout/`**: Core layout management functionality
  - Layout saving, loading, and manipulation logic
  - Data structures for representing window arrangements

- **`sway/`**: Sway window manager integration
  - IPC communication with Sway
  - Window and workspace querying and manipulation

**Note: the IPC communication with sway is intended to be created by us with the objective to remove dependencies on external libraries in the future.**

### `/pkg/`

Library code that can be used by other applications

- **`preset/`**: Layout preset management
  - Preset creation, storage, and retrieval
  - Preset metadata and validation

- **`utils/`**: Common utilities and helper functions
  - Shared utility functions used across the project
  - File operations, string manipulation, etc.

### `/scripts/`

Build scripts, automation, and development tools

- Build automation scripts
- Installation and deployment scripts
- Development environment setup scripts

### `/tests/`

Test files and testing utilities

- **`integration/`**: Integration tests
  - End-to-end testing scenarios
  - Tests that verify component interactions

- **`unit/`**: Unit tests
  - Individual function and method testing
  - Mock objects and isolated testing scenarios

## Directory Conventions

### Internal vs Public Packages

- **`internal/`**: Contains private packages that implement core application logic
- **`pkg/`**: Contains public packages that could be reused by other projects

### Application Entry Points

- **`cmd/`**: Each subdirectory represents a different executable/application
- **`gui/main/`**: Main GUI application entry point

### Documentation Structure

- **`docs/`**: High-level documentation, proposals, and specifications
- **`README.md`**: Project overview and quick start guide

### Testing Organization

- Unit tests should be placed alongside the code they test (Go convention)
- Integration tests are centralized in `/tests/integration/`
- Test utilities and shared test code go in `/tests/`

## Development Guidelines

1. **Import Rules**: Only `/pkg/` packages should be imported by external projects
2. **Configuration**: All config-related code goes in `/internal/config/`
3. **Sway Integration**: All Sway WM interactions should go through `/internal/sway/`
4. **Layout Logic**: Core layout management stays in `/internal/layout/`

This structure follows Go project layout standards and provides clear separation of concerns for the Sway Layout Manager application.