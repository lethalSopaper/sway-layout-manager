package layout

import (
	"strconv"
)

// excludes workspaces matching the given identifiers
func FilterSkipWorkspaces(workspaces []WorkspaceLayout, skipIdentifiers []string) []WorkspaceLayout {
	if len(skipIdentifiers) == 0 {
		return workspaces
	}

	skipNumsMap := make(map[int]bool)
	skipNamesMap := make(map[string]bool)

	// parse identifiers as either numbers or names
	for _, id := range skipIdentifiers {
		if num, err := strconv.Atoi(id); err == nil {
			skipNumsMap[num] = true
		} else {
			skipNamesMap[id] = true
		}
	}

	var filtered []WorkspaceLayout
	for _, ws := range workspaces {

		if skipNumsMap[ws.Num] {
			continue
		}
		if skipNamesMap[ws.Name] {
			continue
		}
		filtered = append(filtered, ws)
	}
	return filtered
}

// includes only workspaces matching the given identifiers
func FilterOnlyWorkspaces(workspaces []WorkspaceLayout, onlyIdentifiers []string) []WorkspaceLayout {
	if len(onlyIdentifiers) == 0 {
		return workspaces
	}

	onlyNumsMap := make(map[int]bool)
	onlyNamesMap := make(map[string]bool)

	// parse identifiers as either numbers or names
	for _, id := range onlyIdentifiers {
		if num, err := strconv.Atoi(id); err == nil {
			onlyNumsMap[num] = true
		} else {
			onlyNamesMap[id] = true
		}
	}

	var filtered []WorkspaceLayout
	for _, ws := range workspaces {
		if onlyNumsMap[ws.Num] || onlyNamesMap[ws.Name] {
			filtered = append(filtered, ws)
		}
	}
	return filtered
}
