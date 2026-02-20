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
