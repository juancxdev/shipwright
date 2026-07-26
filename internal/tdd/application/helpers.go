package application

import (
	"os"
	"path/filepath"
	"time"
)

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func writeFile(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func formatYesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
