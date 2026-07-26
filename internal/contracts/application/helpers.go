package application

import "os"

func artifactExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func formatYesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
