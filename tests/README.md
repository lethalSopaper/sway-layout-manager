# Tests

## Quick Reference

### Running All Tests

```bash
# Complete test suite
go test -v ./tests/...

# Unit tests only
go test -v ./tests/unit/

# Integration tests only
go test -v ./tests/integration/
```

### Running Specific Tests

```bash
# By test name
go test -v ./tests/unit/ -run TestSavePreset
go test -v ./tests/integration/ -run TestSave
go test -v ./tests/... -run TestSway

# By file
go test -v ./tests/unit/preset_test.go
go test -v ./tests/integration/cli_test.go

# Specific test function
go test -v ./tests/unit/ -run TestNewRestorer
go test -v ./tests/integration/ -run TestRestoreBasicFunctionality
```

## Test Categories

### Core Functionality
- **Preset Management**: Save/load/delete/list preset operations
- **Layout Operations**: Parsing, filtering, and transformation
- **Configuration**: XDG compliance and directory management

### Filtering & Operations
- **Workspace Filtering**: Skip/only workspace tests
- **Application Filtering**: Skip/only application tests
- **Floating Windows**: Ignore floating window tests
- **Nested Containers**: Recursive filtering tests

### Restoration
- **Restorer Configuration**: Reuse, clear settings
- **Restore Validation**: Error handling, empty presets
- **Result Tracking**: Workspace/application counters
- **Focus Control**: Post-restoration focus tests

### Integration
- **CLI**: Command-line interface behavior
- **Filesystem**: File I/O and permissions
- **Sway Integration**: Window manager interaction
- **Concurrent Operations**: Multi-threaded access

## Test Environment

- **Sway Version**: v1.11.0
- **Performance**: Sub-20ms test execution
- **Concurrency**: 5 simultaneous operations tested