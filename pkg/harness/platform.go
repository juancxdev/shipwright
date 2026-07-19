package harness

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type PlatformInfo struct {
	OS          string   `json:"os"`
	Arch        string   `json:"arch"`
	IsCI        bool     `json:"ci"`
	HomeDir     string   `json:"home_dir,omitempty"`
	PathEntries []string `json:"path_entries,omitempty"`
}

type SystemProbe interface {
	GOOS() string
	GOARCH() string
	LookPath(binary string) (string, error)
	Stat(path string) (os.FileInfo, error)
	Getenv(key string) string
	UserHomeDir() (string, error)
}

type RealSystemProbe struct{}

func (RealSystemProbe) GOOS() string                           { return runtime.GOOS }
func (RealSystemProbe) GOARCH() string                         { return runtime.GOARCH }
func (RealSystemProbe) LookPath(binary string) (string, error) { return exec.LookPath(binary) }
func (RealSystemProbe) Stat(path string) (os.FileInfo, error)  { return os.Stat(path) }
func (RealSystemProbe) Getenv(key string) string               { return os.Getenv(key) }
func (RealSystemProbe) UserHomeDir() (string, error)           { return os.UserHomeDir() }

func DetectPlatform(probe SystemProbe) PlatformInfo {
	if probe == nil {
		probe = RealSystemProbe{}
	}
	home, _ := probe.UserHomeDir()
	pathValue := probe.Getenv("PATH")
	separator := ":"
	if probe.GOOS() == "windows" {
		separator = ";"
	}
	var pathEntries []string
	if pathValue != "" {
		pathEntries = strings.Split(pathValue, separator)
	}
	return PlatformInfo{
		OS:          probe.GOOS(),
		Arch:        probe.GOARCH(),
		IsCI:        isCI(probe),
		HomeDir:     home,
		PathEntries: pathEntries,
	}
}

func isCI(probe SystemProbe) bool {
	keys := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "BUILDKITE", "CIRCLECI", "TF_BUILD"}
	for _, key := range keys {
		value := strings.ToLower(strings.TrimSpace(probe.Getenv(key)))
		if value != "" && value != "false" && value != "0" {
			return true
		}
	}
	return false
}
