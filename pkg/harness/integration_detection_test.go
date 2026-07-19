package harness

import (
	"errors"
	"os"
	"testing"
	"time"
)

type fakeFileInfo struct {
	name  string
	isDir bool
}

func (f fakeFileInfo) Name() string       { return f.name }
func (f fakeFileInfo) Size() int64        { return 1 }
func (f fakeFileInfo) Mode() os.FileMode  { return 0644 }
func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) IsDir() bool        { return f.isDir }
func (f fakeFileInfo) Sys() any           { return nil }

type fakeProbe struct {
	goos    string
	goarch  string
	home    string
	env     map[string]string
	paths   map[string]string
	statMap map[string]fakeFileInfo
}

func (f fakeProbe) GOOS() string {
	if f.goos == "" {
		return "linux"
	}
	return f.goos
}
func (f fakeProbe) GOARCH() string {
	if f.goarch == "" {
		return "amd64"
	}
	return f.goarch
}
func (f fakeProbe) LookPath(binary string) (string, error) {
	if f.paths != nil {
		if path, ok := f.paths[binary]; ok {
			return path, nil
		}
	}
	return "", errors.New("not found")
}
func (f fakeProbe) Stat(path string) (os.FileInfo, error) {
	if f.statMap != nil {
		if info, ok := f.statMap[path]; ok {
			return info, nil
		}
	}
	return nil, os.ErrNotExist
}
func (f fakeProbe) Getenv(key string) string {
	if f.env != nil {
		return f.env[key]
	}
	return ""
}
func (f fakeProbe) UserHomeDir() (string, error) {
	if f.home == "" {
		return "/home/test", nil
	}
	return f.home, nil
}

func TestDetectPlatformAcrossOS(t *testing.T) {
	tests := []struct {
		name string
		fp   fakeProbe
		want PlatformInfo
	}{
		{
			name: "linux ci",
			fp:   fakeProbe{goos: "linux", goarch: "amd64", env: map[string]string{"CI": "true", "PATH": "/usr/bin:/bin"}},
			want: PlatformInfo{OS: "linux", Arch: "amd64", IsCI: true},
		},
		{
			name: "windows path separator",
			fp:   fakeProbe{goos: "windows", goarch: "amd64", env: map[string]string{"PATH": `C:\bin;D:\tools`}},
			want: PlatformInfo{OS: "windows", Arch: "amd64", IsCI: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectPlatform(tt.fp)
			if got.OS != tt.want.OS || got.Arch != tt.want.Arch || got.IsCI != tt.want.IsCI {
				t.Fatalf("platform = %+v, want %+v", got, tt.want)
			}
			if len(got.PathEntries) == 0 {
				t.Fatal("expected path entries")
			}
		})
	}
}

func TestDetectEngramAcrossPlatforms(t *testing.T) {
	tests := []struct {
		name          string
		probe         fakeProbe
		wantInstalled bool
		wantStatus    string
	}{
		{
			name:          "linux missing falls back",
			probe:         fakeProbe{goos: "linux"},
			wantInstalled: false,
			wantStatus:    DetectionNotInstalled,
		},
		{
			name:          "linux path available",
			probe:         fakeProbe{goos: "linux", paths: map[string]string{"engram": "/usr/local/bin/engram"}},
			wantInstalled: true,
			wantStatus:    DetectionAvailable,
		},
		{
			name:          "windows exe available",
			probe:         fakeProbe{goos: "windows", paths: map[string]string{"engram.exe": `C:\Tools\engram.exe`}},
			wantInstalled: true,
			wantStatus:    DetectionAvailable,
		},
		{
			name:          "env override missing",
			probe:         fakeProbe{goos: "darwin", env: map[string]string{"ENGRAM_BINARY": "/custom/engram"}},
			wantInstalled: false,
			wantStatus:    DetectionNotInstalled,
		},
		{
			name:          "env override available",
			probe:         fakeProbe{goos: "darwin", env: map[string]string{"ENGRAM_BINARY": "/custom/engram"}, statMap: map[string]fakeFileInfo{"/custom/engram": {name: "engram"}}},
			wantInstalled: true,
			wantStatus:    DetectionAvailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectEngram(tt.probe)
			if got.Installed != tt.wantInstalled || got.Status != tt.wantStatus {
				t.Fatalf("DetectEngram = %+v, want installed=%v status=%s", got, tt.wantInstalled, tt.wantStatus)
			}
			if got.Fallback != DecisionsFile {
				t.Fatalf("fallback = %s, want %s", got.Fallback, DecisionsFile)
			}
		})
	}
}

func TestDetectOpenPencilAcrossPlatforms(t *testing.T) {
	macMCP := "/Applications/OpenPencil.app/Contents/Resources/mcp-server.cjs"
	linuxMCP := "/home/dev/.local/share/OpenPencil/resources/mcp-server.cjs"
	windowsMCP := `C:\OpenPencil\mcp-server.cjs`

	tests := []struct {
		name          string
		probe         fakeProbe
		wantInstalled bool
		wantActive    bool
		wantStatus    string
	}{
		{
			name:          "mac app mcp found no canvas",
			probe:         fakeProbe{goos: "darwin", statMap: map[string]fakeFileInfo{macMCP: {name: "mcp-server.cjs"}}},
			wantInstalled: true,
			wantActive:    false,
			wantStatus:    DetectionInstalledNoCanvas,
		},
		{
			name:          "linux mcp found with active canvas env",
			probe:         fakeProbe{goos: "linux", home: "/home/dev", env: map[string]string{"OPENPENCIL_CANVAS_ACTIVE": "true"}, statMap: map[string]fakeFileInfo{linuxMCP: {name: "mcp-server.cjs"}}},
			wantInstalled: true,
			wantActive:    true,
			wantStatus:    DetectionAvailable,
		},
		{
			name:          "windows configured mcp",
			probe:         fakeProbe{goos: "windows", env: map[string]string{"OPENPENCIL_MCP_SERVER": windowsMCP}, statMap: map[string]fakeFileInfo{windowsMCP: {name: "mcp-server.cjs"}}},
			wantInstalled: true,
			wantActive:    false,
			wantStatus:    DetectionInstalledNoCanvas,
		},
		{
			name:          "missing openpencil falls back",
			probe:         fakeProbe{goos: "linux"},
			wantInstalled: false,
			wantActive:    false,
			wantStatus:    DetectionNotInstalled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectOpenPencil(tt.probe)
			if got.Installed != tt.wantInstalled || got.Active != tt.wantActive || got.Status != tt.wantStatus {
				t.Fatalf("DetectOpenPencil = %+v, want installed=%v active=%v status=%s", got, tt.wantInstalled, tt.wantActive, tt.wantStatus)
			}
			if got.Fallback != "design-doc-only" {
				t.Fatalf("fallback = %s", got.Fallback)
			}
		})
	}
}

func TestApplyDetectionStoresPortableMetadata(t *testing.T) {
	integrations := DefaultIntegrations()
	engram := DetectEngram(fakeProbe{goos: "linux", goarch: "arm64", paths: map[string]string{"engram": "/usr/bin/engram"}})
	openpencil := DetectOpenPencil(fakeProbe{goos: "linux"})

	integrations.ApplyDetection(engram, openpencil)

	if integrations.Platform.OS != "linux" || integrations.Platform.Arch != "arm64" {
		t.Fatalf("platform = %+v", integrations.Platform)
	}
	if integrations.Engram.BinaryPath != "/usr/bin/engram" {
		t.Fatalf("engram binary = %s", integrations.Engram.BinaryPath)
	}
	if integrations.OpenPencil.Status != DetectionNotInstalled {
		t.Fatalf("openpencil status = %s", integrations.OpenPencil.Status)
	}
	if integrations.Engram.LastDetectedAt == "" || integrations.OpenPencil.LastDetectedAt == "" {
		t.Fatal("expected last_detected_at fields")
	}
}

func TestOpenPencilAppPathStoredSeparatelyFromMCP(t *testing.T) {
	appPath := "/Applications/OpenPencil.app"
	openpencil := DetectOpenPencil(fakeProbe{
		goos: "darwin",
		env:  map[string]string{"OPENPENCIL_APP_PATH": appPath},
		statMap: map[string]fakeFileInfo{
			appPath: {name: "OpenPencil.app", isDir: true},
		},
	})

	if openpencil.PathKind != DetectionPathApp {
		t.Fatalf("path kind = %s, want %s", openpencil.PathKind, DetectionPathApp)
	}

	integrations := DefaultIntegrations()
	integrations.ApplyDetection(DetectEngram(fakeProbe{goos: "darwin"}), openpencil)

	if integrations.OpenPencil.AppPath != appPath {
		t.Fatalf("app path = %s, want %s", integrations.OpenPencil.AppPath, appPath)
	}
	if integrations.OpenPencil.MCPServerPath != "" {
		t.Fatalf("mcp server path should stay empty, got %s", integrations.OpenPencil.MCPServerPath)
	}
}

func TestDetectOpenPencilMCPCommandFromPath(t *testing.T) {
	result := DetectOpenPencilWithConfig(fakeProbe{
		goos:  "darwin",
		paths: map[string]string{"openpencil-mcp": "/opt/homebrew/bin/openpencil-mcp"},
	}, DefaultPortableConfig())

	if !result.Installed || result.PathKind != DetectionPathBinary || result.Path != "/opt/homebrew/bin/openpencil-mcp" {
		t.Fatalf("openpencil command detection = %+v", result)
	}
	if result.Status != DetectionInstalledNoCanvas {
		t.Fatalf("status = %s", result.Status)
	}
}
