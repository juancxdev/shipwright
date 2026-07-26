package application

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	projectprofile "shipwright/internal/projectprofile/application"
)

type ProjectProfile = projectprofile.ProjectProfile
type StackSignal = projectprofile.StackSignal
type RepositoryProfile = projectprofile.RepositoryProfile
type ProjectStructure = projectprofile.ProjectStructure
type ProjectCommands = projectprofile.ProjectCommands
type DetectedCommand = projectprofile.DetectedCommand
type TDDCapability = projectprofile.TDDCapability

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func writeFile(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func artifactExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func sortedUnique(values []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		result = append(result, trimmed)
	}
	sort.Strings(result)
	return result
}

func uniqueStacks(values []StackSignal) []StackSignal {
	seen := map[string]bool{}
	var result []StackSignal
	for _, value := range values {
		key := strings.ToLower(value.Name + "|" + value.Kind)
		if strings.TrimSpace(value.Name) == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name == result[j].Name {
			return result[i].Kind < result[j].Kind
		}
		return result[i].Name < result[j].Name
	})
	return result
}
