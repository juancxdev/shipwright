package application

import (
	"errors"
	"os"
	"testing"
	"time"

	config "shipwright/internal/config/application"
)

type autoConfigFakeFileInfo struct {
	name  string
	isDir bool
}

func (f autoConfigFakeFileInfo) Name() string       { return f.name }
func (f autoConfigFakeFileInfo) Size() int64        { return 1 }
func (f autoConfigFakeFileInfo) Mode() os.FileMode  { return 0o644 }
func (f autoConfigFakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f autoConfigFakeFileInfo) IsDir() bool        { return f.isDir }
func (f autoConfigFakeFileInfo) Sys() any           { return nil }

type autoConfigFakeProbe struct {
	goos    string
	goarch  string
	home    string
	env     map[string]string
	paths   map[string]string
	statMap map[string]autoConfigFakeFileInfo
}

func (f autoConfigFakeProbe) GOOS() string {
	if f.goos == "" {
		return "linux"
	}
	return f.goos
}
func (f autoConfigFakeProbe) GOARCH() string {
	if f.goarch == "" {
		return "amd64"
	}
	return f.goarch
}
func (f autoConfigFakeProbe) LookPath(binary string) (string, error) {
	if f.paths != nil {
		if path, ok := f.paths[binary]; ok {
			return path, nil
		}
	}
	return "", errors.New("not found")
}
func (f autoConfigFakeProbe) Stat(path string) (os.FileInfo, error) {
	if f.statMap != nil {
		if info, ok := f.statMap[path]; ok {
			return info, nil
		}
	}
	return nil, os.ErrNotExist
}
func (f autoConfigFakeProbe) Getenv(key string) string {
	if f.env != nil {
		return f.env[key]
	}
	return ""
}
func (f autoConfigFakeProbe) UserHomeDir() (string, error) {
	if f.home == "" {
		return "/home/test", nil
	}
	return f.home, nil
}

func TestAutoConfigureOpenDesignUsesODFromPath(t *testing.T) {
	cfg := config.DefaultPortableConfig()
	got := AutoConfigureOpenDesign(autoConfigFakeProbe{
		goos:  "darwin",
		paths: map[string]string{"od": "/opt/homebrew/bin/od"},
		statMap: map[string]autoConfigFakeFileInfo{
			"/opt/homebrew/bin/od":                     {name: "od"},
			"/tmp/open-design/ipc/default/daemon.sock": {name: "daemon.sock"},
		},
	}, cfg)

	if cfg.Integrations.OpenDesign.MCPCommand != "/opt/homebrew/bin/od" {
		t.Fatalf("command = %q", cfg.Integrations.OpenDesign.MCPCommand)
	}
	if len(cfg.Integrations.OpenDesign.MCPArgs) != 1 || cfg.Integrations.OpenDesign.MCPArgs[0] != "mcp" {
		t.Fatalf("args = %+v, want [mcp]", cfg.Integrations.OpenDesign.MCPArgs)
	}
	if got.Status != DetectionAvailable || !got.Available {
		t.Fatalf("detection = %+v, want available", got)
	}
}

func TestAutoConfigureOpenDesignUsesDaemonInstallInfoFirst(t *testing.T) {
	originalFetch := fetchOpenDesignInstallInfo
	t.Cleanup(func() { fetchOpenDesignInstallInfo = originalFetch })

	fetchOpenDesignInstallInfo = func(baseURL string) (*openDesignInstallInfo, error) {
		if baseURL != "http://127.0.0.1:56868" {
			t.Fatalf("baseURL = %q, want daemon URL from environment", baseURL)
		}
		return &openDesignInstallInfo{
			Command:    "/opt/homebrew/bin/node",
			Args:       []string{"/tools/open-design/apps/daemon/dist/cli.js", "mcp"},
			Env:        map[string]string{"OD_DATA_DIR": "/tools/open-design/.od", "OD_SIDECAR_IPC_PATH": "/tmp/open-design/ipc/default/daemon.sock"},
			DaemonURL:  "http://127.0.0.1:56868",
			WebBaseURL: "http://127.0.0.1:56869",
			CliExists:  true,
			NodeExists: true,
		}, nil
	}

	cfg := config.DefaultPortableConfig()
	got := AutoConfigureOpenDesign(autoConfigFakeProbe{
		goos: "darwin",
		env:  map[string]string{"OD_DAEMON_URL": "http://127.0.0.1:56868"},
		paths: map[string]string{
			"od": "/opt/homebrew/bin/od",
		},
		statMap: map[string]autoConfigFakeFileInfo{
			"/opt/homebrew/bin/node":                   {name: "node"},
			"/tmp/open-design/ipc/default/daemon.sock": {name: "daemon.sock"},
		},
	}, cfg)

	if cfg.Integrations.OpenDesign.MCPCommand != "/opt/homebrew/bin/node" {
		t.Fatalf("command = %q", cfg.Integrations.OpenDesign.MCPCommand)
	}
	wantArgs := []string{"/tools/open-design/apps/daemon/dist/cli.js", "mcp"}
	if len(cfg.Integrations.OpenDesign.MCPArgs) != len(wantArgs) {
		t.Fatalf("args = %+v, want %+v", cfg.Integrations.OpenDesign.MCPArgs, wantArgs)
	}
	for i := range wantArgs {
		if cfg.Integrations.OpenDesign.MCPArgs[i] != wantArgs[i] {
			t.Fatalf("args = %+v, want %+v", cfg.Integrations.OpenDesign.MCPArgs, wantArgs)
		}
	}
	if cfg.Integrations.OpenDesign.DataDir != "/tools/open-design/.od" {
		t.Fatalf("data dir = %q", cfg.Integrations.OpenDesign.DataDir)
	}
	if cfg.Integrations.OpenDesign.DaemonURL != "http://127.0.0.1:56868" {
		t.Fatalf("daemon URL = %q", cfg.Integrations.OpenDesign.DaemonURL)
	}
	if cfg.Integrations.OpenDesign.IPCPath != "/tmp/open-design/ipc/default/daemon.sock" {
		t.Fatalf("ipc path = %q", cfg.Integrations.OpenDesign.IPCPath)
	}
	if got.DaemonURL != "http://127.0.0.1:56868" {
		t.Fatalf("detection daemon URL = %q", got.DaemonURL)
	}
	if got.Status != DetectionAvailable || !got.Available {
		t.Fatalf("detection = %+v, want available", got)
	}
}

func TestAutoConfigureOpenDesignFindsLocalCheckoutAndNode(t *testing.T) {
	root := "/Users/dev/Documents/TOOLS/APPs/open-design"
	cli := root + "/apps/daemon/dist/cli.js"
	cfg := config.DefaultPortableConfig()
	got := AutoConfigureOpenDesign(autoConfigFakeProbe{
		goos: "darwin",
		home: "/Users/dev",
		paths: map[string]string{
			"node": "/opt/homebrew/bin/node",
		},
		statMap: map[string]autoConfigFakeFileInfo{
			cli:                      {name: "cli.js"},
			"/opt/homebrew/bin/node": {name: "node"},
			"/tmp/open-design/ipc/default/daemon.sock": {name: "daemon.sock"},
		},
	}, cfg)

	if cfg.Integrations.OpenDesign.MCPCommand != "/opt/homebrew/bin/node" {
		t.Fatalf("command = %q", cfg.Integrations.OpenDesign.MCPCommand)
	}
	wantArgs := []string{cli, "mcp"}
	if len(cfg.Integrations.OpenDesign.MCPArgs) != len(wantArgs) {
		t.Fatalf("args = %+v, want %+v", cfg.Integrations.OpenDesign.MCPArgs, wantArgs)
	}
	for i := range wantArgs {
		if cfg.Integrations.OpenDesign.MCPArgs[i] != wantArgs[i] {
			t.Fatalf("args = %+v, want %+v", cfg.Integrations.OpenDesign.MCPArgs, wantArgs)
		}
	}
	if cfg.Integrations.OpenDesign.DataDir != root+"/.od" {
		t.Fatalf("data dir = %q", cfg.Integrations.OpenDesign.DataDir)
	}
	if got.Status != DetectionAvailable || !got.Available {
		t.Fatalf("detection = %+v, want available", got)
	}
}

func TestAutoConfigureOpenDesignFallsBackWhenNothingDetected(t *testing.T) {
	cfg := config.DefaultPortableConfig()
	got := AutoConfigureOpenDesign(autoConfigFakeProbe{}, cfg)

	if cfg.Integrations.OpenDesign.MCPCommand != "" || len(cfg.Integrations.OpenDesign.MCPArgs) != 0 {
		t.Fatalf("config should remain empty, got command=%q args=%+v", cfg.Integrations.OpenDesign.MCPCommand, cfg.Integrations.OpenDesign.MCPArgs)
	}
	if got.Status != DetectionNotInstalled {
		t.Fatalf("status = %s, want %s", got.Status, DetectionNotInstalled)
	}
}
