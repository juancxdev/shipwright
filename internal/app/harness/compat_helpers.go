package harness

import (
	"os"
	"sort"
	"strings"
)

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileContainsAny(path string, needles []string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	content := strings.ToLower(string(data))
	for _, needle := range needles {
		if strings.Contains(content, strings.ToLower(needle)) {
			return true
		}
	}
	return false
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

func firstOrDefault(values []string, fallback string) string {
	if len(values) == 0 {
		return fallback
	}
	return values[0]
}

func defaultSlice(values []string, fallback []string) []string {
	if len(values) == 0 {
		return fallback
	}
	return values
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func trimEnv(value string) string {
	return strings.TrimSpace(value)
}
