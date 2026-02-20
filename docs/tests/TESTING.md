# Testing Documentation

## Overview

This document provides comprehensive documentation for the Sway Layout Manager test suite. The test suite covers both **unit tests** and **integration tests** across multiple test files.
For test execution commands and quick reference, see [README.md](README.md).

### Unit Tests

Unit tests are focused on testing **individual components in isolation**. They validate that each module works correctly independent of external dependencies like file systems, network connections, or external processes.

### Integration Tests

Integration tests validate **end-to-end functionality** by testing how components work together in real environments. They verify that the entire system operates correctly with actual dependencies.

---

## Unit Tests

**Directory**: `tests/unit/`

### Configuration Tests (`config_test.go`)

#### `TestNewConfig`

-   **Validates**: Configuration constructor functionality and XDG Base Directory specification compliance
-   **Tests**:
    -   `creates_config_with_xdg_variables` - Uses XDG_CONFIG_HOME and XDG_DATA_HOME when set
    -   `falls_back_to_home_directory` - Uses ~/.config and ~/.local/share when XDG vars unset
    -   `creates_directories_with_correct_permissions` - Ensures 0755 permissions for created directories
    -   `handles_permission_errors` - Handling when directories can't be created

#### `TestConfigPresetPath`

-   **Validates**: Preset file path generation functionality and file path construction with JSON extension handling
-   **Tests**:
    -   `generates_correct_path_with_extension` - Builds proper file paths with .json extension
    -   `adds_json_extension_when_missing` - Auto-appends .json to preset names
    -   `handles_special_characters` - Properly escapes special characters in names
    -   `validates_path_structure` - Ensures paths follow expected directory structure

#### `TestConfigDefaultPresetName`

-   **Validates**: Default preset name generation with timestamp formatting and auto-generated preset names for save operations
-   **Tests**:
    -   `generates_timestamp_based_names` - Creates names like "layout-2026-01-03-12-30-45"
    -   `formats_timestamp_correctly` - Uses proper ISO-like date format
    -   `ensures_unique_names` - Different calls generate different timestamps
    -   `validates_name_pattern` - Follows "layout-YYYY-MM-DD-HH-MM-SS" pattern

### Layout Processing Tests (`layout_test.go`)

#### `TestLayoutTypes`

-   **Validates**: Layout data structure definitions and JSON serialization for data structures
-   **Tests**:
    -   `layout_struct_serialization` - Proper JSON marshaling/unmarshaling of Layout struct
    -   `workspace_struct_serialization` - Workspace data structure handling
    -   `container_struct_serialization` - Container hierarchy serialization
    -   `node_struct_serialization` - Individual node data handling

#### `TestNewParser`

-   **Validates**: Layout parser initialization and configuration with different client configurations
-   **Tests**:
    -   `creates_parser_with_valid_client` - Successful parser creation with Sway client
    -   `handles_nil_client` - Handling of nil client parameter
    -   `initializes_parser_state` - Proper initial state setup
    -   `validates_parser_configuration` - Configuration validation and error handling

#### `TestCaptureCurrentLayout`

-   **Validates**: Live workspace layout capture from Sway and real-time layout data extraction from Sway window manager
-   **Tests**:
    -   `captures_full_workspace_tree` - Complete workspace hierarchy capture
    -   `handles_empty_workspaces` - Proper handling of workspaces without containers
    -   `processes_multiple_outputs` - Multi-monitor layout capture
    -   `handles_sway_connection_errors` - Degradation when Sway unavailable

### Preset Management Tests (`preset_test.go`)

#### `TestNewManager`

-   **Validates**: Preset manager initialization and manager creation with configuration dependencies
-   **Tests**:
    -   `creates_manager_with_valid_config` - Successful manager creation
    -   `validates_config_dependency` - Proper config validation
    -   `initializes_manager_state` - Initial state setup
    -   `handles_invalid_config` - Error handling for invalid configurations

#### `TestSavePreset`

-   **Validates**: Core preset saving functionality, complete save operation from layout data to file creation
-   **Tests**:
    -   `saves_valid_preset_data` - Successful save with valid layout data
    -   `creates_json_file_with_correct_format` - Proper JSON file generation
    -   `handles_existing_file_overwrite` - Overwrites existing presets safely
    -   `validates_preset_name` - Name validation and sanitization
    -   `sets_correct_file_permissions` - Files created with 0644 permissions
    -   `handles_write_permission_errors` - Error handling for permission issues

#### `TestLoadPreset`

-   **Validates**: Preset loading and validation functionality with file reading, JSON parsing, and data validation
-   **Tests**:
    -   `loads_valid_preset_file` - Successful loading of saved presets
    -   `validates_json_format` - Proper JSON structure validation
    -   `handles_missing_files` - Error handling for non-existent presets
    -   `handles_corrupted_json` - Handling of invalid JSON files
    -   `validates_preset_data_structure` - Data structure integrity checking

#### `TestDeletePreset`

-   **Validates**: Preset deletion functionality and file removal with cleanup operations
-   **Tests**:
    -   `deletes_existing_preset` - Successful file removal
    -   `handles_non_existent_files` - Error handling for missing files
    -   `confirms_file_removal` - Verification that files are actually deleted
    -   `handles_permission_errors` - Error handling for permission issues

#### `TestExistsPreset`

-   **Validates**: Preset existence checking functionality and file existence verification without loading content
-   **Tests**:
    -   `detects_existing_presets` - Correctly identifies existing preset files
    -   `detects_missing_presets` - Correctly identifies non-existent presets
    -   `handles_permission_issues` - Handles files that exist but aren't readable
    -   `validates_file_vs_directory` - Distinguishes between files and directories

#### `TestListPresets`

-   **Validates**: Preset enumeration and listing functionality with directory scanning and preset file discovery
-   **Tests**:
    -   `lists_all_available_presets` - Complete preset file enumeration
    -   `filters_json_files_only` - Only includes .json preset files
    -   `handles_empty_directories` - Handling of empty preset directories
    -   `sorts_presets_alphabetically` - Proper preset name sorting
    -   `handles_directory_permission_errors` - Error handling for unreadable directories

### Sway Tests (`tests/unit/sway_test.go`)

#### `TestNewClient`

-   **Validates**: Sway client creation and initialization with connection setup to Sway window manager
-   **Tests**:
    -   `creates_client_when_sway_available` - Successful client creation with swaymsg
    -   `fails_gracefully_when_sway_unavailable` - Error handling when Sway not running
    -   `validates_swaymsg_command` - Verification that swaymsg is in PATH
    -   `handles_invalid_sway_socket` - Error handling for invalid socket connections

#### `TestClientMethods`

-   **Validates**: Sway client method functionality and Sway IPC command execution with response handling
-   **Tests**:
    -   `GetVersion_returns_sway_version` - Retrieves Sway version information
    -   `GetTree_returns_workspace_tree` - Fetches complete workspace hierarchy
    -   `GetWorkspaces_returns_workspace_list` - Gets list of all workspaces
    -   `handles_command_execution_errors` - Error handling for failed commands
    -   `validates_json_response_parsing` - Proper parsing of Sway JSON responses

#### `TestSwayTypes`

-   **Validates**: Sway data structure definitions and JSON compatibility for Sway IPC responses
-   **Tests**:
    -   `Node_JSON_unmarshaling` - Node structure JSON compatibility
    -   `Workspace_JSON_unmarshaling` - Workspace structure parsing
    -   `Version_JSON_unmarshaling` - Version information parsing
    -   `validates_field_mappings` - Proper field mapping from Sway JSON
    -   `handles_optional_fields` - Handling of optional JSON fields

---

## Integration Tests

**Directory**: `tests/integration/`

### CLI Tests (`cli_test.go`)

#### `TestNewClient`

-   **Validates**: Sway client creation and initialization with connection setup to Sway window manager
-   **Tests**:
    -   `creates_client_when_sway_available` - Successful client creation with swaymsg
    -   `fails_gracefully_when_sway_unavailable` - Error handling when Sway not running
    -   `validates_swaymsg_command` - Verification that swaymsg is in PATH
    -   `handles_invalid_sway_socket` - Error handling for invalid socket connections

#### `TestClientMethods`

-   **Validates**: Sway client method functionality and Sway IPC command execution with response handling
-   **Tests**:
    -   `GetVersion_returns_sway_version` - Retrieves Sway version information
    -   `GetTree_returns_workspace_tree` - Fetches complete workspace hierarchy
    -   `GetWorkspaces_returns_workspace_list` - Gets list of all workspaces
    -   `handles_command_execution_errors` - Error handling for failed commands
    -   `validates_json_response_parsing` - Proper parsing of Sway JSON responses

#### `TestSwayTypes`

-   **Validates**: Sway data structure definitions and JSON compatibility for Sway IPC responses
-   **Tests**:
    -   `Node_JSON_unmarshaling` - Node structure JSON compatibility
    -   `Workspace_JSON_unmarshaling` - Workspace structure parsing
    -   `Version_JSON_unmarshaling` - Version information parsing
    -   `validates_field_mappings` - Proper field mapping from Sway JSON
    -   `handles_optional_fields` - Handling of optional JSON fields

#### `TestSwayAvailability`

-   **Validates**: Sway window manager detection and availability with Sway environment detection and graceful degradation
-   **Tests**:
    -   `swaymsg_command_availability` - Tests swaymsg detection in PATH
    -   `sway_client_creation_without_swaymsg` - Tests failure when Sway unavailable

#### `TestSwayConnectionAttempts`

-   **Validates**: Live connection to running Sway instance. Real Sway v1.11.0 integration and data retrieval
-   **Tests**:
    -   `get_sway_version` - Retrieves live Sway version (validates v1.11.0)
    -   `get_sway_tree` - Fetches complete workspace tree from live environment
    -   `get_sway_workspaces` - Gets workspace list (validates 3-4 workspaces including custom like "AUDIO")

#### `TestSwayJSONParsing`

-   **Validates**: JSON parsing of complex Sway data structures and complex JSON response handling from Sway IPC
-   **Tests**:
    -   `parse_sway_version_JSON` - Version information JSON parsing
    -   `parse_sway_workspace_JSON` - Workspace data structure parsing
    -   `parse_complex_sway_tree_JSON` - Complex tree with idle_inhibitors and nested structures

#### `TestSwayCommandTimeout`

-   **Validates**: Sway command execution performance and timeout handling with performance characteristics and timeout management
-   **Tests**:
    -   `version_command_completes_quickly` - Version command performance (validates <10ms)
    -   `tree_command_completes_quickly` - Tree command performance (validates <10ms)

#### `TestSwayErrorHandling`

-   **Validates**: Error handling and degradation for Sway operations with robustness and error resilience
-   **Tests**:
    -   `handles_sway_not_running` - Handling when Sway is not available

#### `TestHelpFlag`

-   **Validates**: Help flag functionality in the compiled binary and help system with usage information display
-   **Tests**:
    -   `long_help_flag` - Tests --help flag behavior
    -   `short_help_flag` - Tests -h flag behavior

#### `TestVersionFlag`

-   **Validates**: Version flag functionality in the compiled binary and version information display with formatting
-   **Tests**:
    -   `long_version_flag` - Tests --version flag behavior
    -   `short_version_flag` - Tests -v flag behavior

#### `TestInvalidCommand`

-   **Validates**: Error handling for invalid CLI commands and error messages with graceful failure handling
-   **Tests**:
    -   `invalid_command` - Tests response to unknown commands
    -   `typo_in_command` - Tests response to misspelled commands
    -   `empty_args` - Tests response to no arguments provided

#### `TestListCommandEmpty`

-   **Validates**: List command when no presets exist and empty state handling with user guidance
-   **Tests**: List command behavior with empty preset directory

#### `TestSaveCommandInvalidName`

-   **Validates**: Save command with various invalid input scenarios - **CRITICAL SAVE TESTING** - save command robustness and auto-naming functionality
-   **Tests**:
    -   `empty_name` - Tests save with empty preset name (auto-generates name)
    -   `whitespace_only` - Tests save with whitespace-only names
    -   `special_characters` - Tests save with special characters in names

#### `TestDeleteCommandNonExistent`

-   **Validates**: Delete command with non-existent presets and error handling for delete operations on missing files
-   **Tests**: Delete command behavior with non-existent preset names

#### `TestBinaryExecution`

-   **Validates**: Binary execution characteristics and process management with binary behavior, performance, and signal handling
-   **Tests**:
    -   `binary_exits_quickly_for_help` - Performance testing for help command
    -   `binary_handles_signals_gracefully` - Signal handling and cleanup

### Configuration Tests (`config_test.go`)

#### `TestXDGEnvironmentVariables`

-   **Validates**: XDG Base Directory Specification compliance in real environments and Linux desktop standards compliance with environment variable handling
-   **Tests**:
    -   `respects_XDG_CONFIG_HOME` - Uses custom config directories when XDG_CONFIG_HOME set
    -   `respects_XDG_DATA_HOME` - Uses custom data directories when XDG_DATA_HOME set
    -   `falls_back_to_home_directory` - Falls back to ~/.config and ~/.local/share appropriately

#### `TestConfigurationPersistence`

-   **Validates**: Configuration loading and directory creation in real environments with configuration system integration with filesystem
-   **Tests**:
    -   `creates_directories_correctly` - Creates proper directory structure
    -   `succeeds_without_existing_files` - Works without pre-existing configuration

#### `TestDefaultConfiguration`

-   **Validates**: Default configuration behavior and sensible defaults with default value provision and system initialization
-   **Tests**:
    -   `provides_default_values` - Sets appropriate default values
    -   `creates_necessary_directories` - Creates required directories with correct permissions

#### `TestConfigurationDirectoryStructure`

-   **Validates**: Complete application directory structure creation with full directory hierarchy and permission management
-   **Tests**:
    -   `creates_complete_directory_structure` - Creates full XDG directory hierarchy
    -   `directory_permissions_are_correct` - Sets proper directory permissions (0755)

#### `TestConfigurationIsolation`

-   **Validates**: Configuration isolation between different environments and multi-user with multi-environment configuration handling
-   **Tests**:
    -   `multiple_configs_with_different_environments` - Ensures different configs don't interfere

### Filesystem Tests (`filesystem_test.go`)

#### `TestXDGDirectoryCreation`

-   **Validates**: Real XDG directory creation and structure validation with Linux desktop directory standards compliance
-   **Tests**: XDG Base Directory structure creation with proper hierarchy

#### `TestPresetFilePersistence`

-   **Validates**: Complete preset save/load cycles with real files, end-to-end preset persistence functionality
-   **Tests**: Complex preset data persistence through complete save/load cycles

#### `TestFilePermissions`

-   **Validates**: Unix file permission handling for security and access control with proper file and directory permissions for multi-user systems
-   **Tests**: File permissions (644) and directory permissions (755) validation

#### `TestJSONFormatValidation`

-   **Validates**: JSON format validation and complex data structure handling with JSON serialization/deserialization with complex data
-   **Tests**: Complex JSON preset validation with 1406-byte files containing full workspace layouts

#### `TestConcurrentAccess`

-   **Validates**: Concurrent file operations and thread safety with multi-user scenarios and concurrent save operations
-   **Tests**: 5 simultaneous preset save operations to test concurrency handling

#### `TestDeleteFileRemoval`

-   **Validates**: Complete file deletion and cleanup operations with file removal operations and cleanup verification
-   **Tests**: File deletion with verification of complete removal

### Sway Environment Tests (`tests/integration/sway_test.go`)

#### `TestSwayAvailability`

-   **Validates**: Sway window manager detection and availability with Sway environment detection and graceful degradation
-   **Tests**:
    -   `swaymsg_command_availability` - Tests swaymsg detection in PATH
    -   `sway_client_creation_without_swaymsg` - Tests failure when Sway unavailable

#### `TestSwayConnectionAttempts`

-   **Validates**: Live connection to running Sway instance. Real Sway v1.11.0 integration and data retrieval
-   **Tests**:
    -   `get_sway_version` - Retrieves live Sway version (validates v1.11.0)
    -   `get_sway_tree` - Fetches complete workspace tree from live environment
    -   `get_sway_workspaces` - Gets workspace list (validates 3-4 workspaces including custom like "AUDIO")

#### `TestSwayJSONParsing`

-   **Validates**: JSON parsing of complex Sway data structures and complex JSON response handling from Sway IPC
-   **Tests**:
    -   `parse_sway_version_JSON` - Version information JSON parsing
    -   `parse_sway_workspace_JSON` - Workspace data structure parsing
    -   `parse_complex_sway_tree_JSON` - Complex tree with idle_inhibitors and nested structures

#### `TestSwayCommandTimeout`

-   **Validates**: Sway command execution performance and timeout handling with performance characteristics and timeout management
-   **Tests**:
    -   `version_command_completes_quickly` - Version command performance (validates <10ms)
    -   `tree_command_completes_quickly` - Tree command performance (validates <10ms)

#### `TestSwayErrorHandling`

-   **Validates**: Error handling and degradation for Sway operations with robustness and error resilience
-   **Tests**:
    -   `handles_sway_not_running` - Handling when Sway is not available

---

## Test Environment Requirements

### System Requirements

-   **Go 1.25.5+**: For testing framework and language features
-   **Linux OS**: For XDG Base Directory Specification compliance
-   **Sway v1.11.0** (optional): For integration testing
-   **Unix permissions**: For file permission testing (644/755)

### Environment Variables

-   `XDG_CONFIG_HOME`: Custom configuration directory for testing
-   `XDG_DATA_HOME`: Custom data directory for testing
-   `PATH`: Must include `swaymsg` for Sway integration tests

---

## Success Metrics

-   [x] **All tests passing** - Complete test coverage
-   [x] **Real Sway v1.11.0 integration** - Live environment validation
-   [x] **Sub-10ms performance** - High-performance Sway operations
-   [x] **1406-byte JSON files** - Complex data structure handling
-   [x] **XDG compliance** - Linux desktop standards adherence
-   [x] **Concurrent operations** - 5 simultaneous save operations
-   [x] **Security compliance** - Proper Unix permissions (644/755)
-   [x] **Graceful degradation** - Works without Sway environment
