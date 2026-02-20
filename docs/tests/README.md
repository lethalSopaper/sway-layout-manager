# Tests Documentation

This directory contains comprehensive documentation related to testing for Sway Layout Manager.

## Documentation Files

- **[TESTING.md](TESTING.md)** - Comprehensive testing documentation
    -   Unit tests - Isolated component testing
    -   Integration tests - End-to-end functionality testing
    -   Execution instructions and troubleshooting
    -   Detailed explanations for each test function

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
go test -v ./tests/unit/ -run "TestSavePreset"
go test -v ./tests/integration/ -run "TestSave"
go test -v ./tests/... -run "TestSway"

# By file
go test -v ./tests/unit/preset_test.go
go test -v ./tests/integration/cli_test.go
```

### Test Categories

-   **Save Functionality**: Core preset save/load operations
-   **CLI**: Command-line interface behavior
-   **Filesystem Operations**: File I/O and permissions
-   **Sway Integration**: Window manager interaction
-   **Configuration**: XDG compliance and directory management

### Test Environment

-   **Real Sway v1.11.0** integration
-   **Sub-10ms performance** validation
-   **1406-byte JSON** complex data handling
-   **Concurrent operations** testing 5 simultaneous

For complete documentation, see [TESTING.md](TESTING.md).