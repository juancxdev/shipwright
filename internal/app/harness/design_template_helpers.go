package harness

import "strings"

func requestOrProjectName(state *State, request string) string {
	if strings.TrimSpace(request) != "" {
		return request
	}
	if state != nil && strings.TrimSpace(state.ProjectName) != "" {
		return state.ProjectName
	}
	return "(not set)"
}
