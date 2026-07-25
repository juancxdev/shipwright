package cmd

import (
	"strings"
	"testing"

	"shipwright/pkg/harness"
)

func TestParseInitOptionsDefaultsToInteractiveOpenCode(t *testing.T) {
	options := parseInitOptions(nil)
	if options.Executor != harness.ExecutorOpenCode {
		t.Fatalf("executor = %s", options.Executor)
	}
	if !options.IntegrationWizard {
		t.Fatal("expected integration wizard enabled by default")
	}
}

func TestParseInitOptionsCanDisableIntegrationWizard(t *testing.T) {
	options := parseInitOptions([]string{"--no-interactive"})
	if options.IntegrationWizard {
		t.Fatal("expected integration wizard disabled")
	}

	options = parseInitOptions([]string{"--no-integrations-wizard"})
	if options.IntegrationWizard {
		t.Fatal("expected integration wizard disabled")
	}
}

func TestParseInitIntegrationSelection(t *testing.T) {
	options := initIntegrationOptions(
		harness.DetectionResult{Status: harness.DetectionAvailable, Available: true},
		harness.DetectionResult{Status: harness.DetectionNotInstalled},
		harness.DetectionResult{Status: harness.DetectionConfiguredUnverified, Configured: true},
		harness.DetectionResult{Status: harness.DetectionInstalledNoCanvas, Installed: true},
	)

	tests := []struct {
		name  string
		input string
		want  map[string]bool
	}{
		{
			name:  "enter uses recommended",
			input: "",
			want:  map[string]bool{"engram": true, "stitch": true},
		},
		{
			name:  "numbers",
			input: "1,2",
			want:  map[string]bool{"engram": true, "stitch": true},
		},
		{
			name:  "names",
			input: "engram openpencil",
			want:  map[string]bool{"engram": true, "openpencil": true},
		},
		{
			name:  "none",
			input: "0",
			want:  map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInitIntegrationSelection(tt.input, options)
			if err != nil {
				t.Fatalf("parseInitIntegrationSelection error = %v", err)
			}
			assertSelectedIntegrations(t, got, tt.want)
		})
	}
}

func TestParseInitIntegrationSelectionRejectsUnknown(t *testing.T) {
	options := initIntegrationOptions(harness.DetectionResult{}, harness.DetectionResult{}, harness.DetectionResult{}, harness.DetectionResult{})
	if _, err := parseInitIntegrationSelection("9", options); err == nil {
		t.Fatal("expected error for unknown integration number")
	}
	if _, err := parseInitIntegrationSelection("figma", options); err == nil {
		t.Fatal("expected error for unknown integration name")
	}
}

func TestInitIntegrationSelectorStateMoveAndToggle(t *testing.T) {
	options := initIntegrationOptions(
		harness.DetectionResult{Status: harness.DetectionAvailable, Available: true},
		harness.DetectionResult{Status: harness.DetectionNotInstalled},
		harness.DetectionResult{Status: harness.DetectionConfiguredUnverified, Configured: true},
		harness.DetectionResult{Status: harness.DetectionInstalledNoCanvas, Installed: true},
	)
	state := newInitIntegrationSelectorState(options)

	assertSelectedIntegrations(t, state.CloneSelected(), map[string]bool{"engram": true, "stitch": true})

	state.Move(1)
	if state.Cursor != 1 {
		t.Fatalf("cursor = %d, want 1", state.Cursor)
	}
	state.Toggle()
	assertSelectedIntegrations(t, state.CloneSelected(), map[string]bool{"engram": true})

	state.Move(-1)
	if state.Cursor != 0 {
		t.Fatalf("cursor = %d, want 0", state.Cursor)
	}
	state.Move(-1)
	if state.Cursor != 3 {
		t.Fatalf("cursor wrapped = %d, want 3", state.Cursor)
	}
	state.Toggle()
	assertSelectedIntegrations(t, state.CloneSelected(), map[string]bool{"engram": true, "openpencil": true})
}

func TestReadInitSelectorKey(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want initSelectorKey
	}{
		{name: "j down", in: "j", want: initSelectorKeyDown},
		{name: "k up", in: "k", want: initSelectorKeyUp},
		{name: "space toggle", in: " ", want: initSelectorKeyToggle},
		{name: "enter submit", in: "\n", want: initSelectorKeySubmit},
		{name: "q cancel", in: "q", want: initSelectorKeyCancel},
		{name: "arrow up", in: "\x1b[A", want: initSelectorKeyUp},
		{name: "arrow down", in: "\x1b[B", want: initSelectorKeyDown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readInitSelectorKey(strings.NewReader(tt.in))
			if err != nil {
				t.Fatalf("readInitSelectorKey error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("key = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderInitIntegrationSelectorViewUsesRawSafeLineEndings(t *testing.T) {
	options := initIntegrationOptions(
		harness.DetectionResult{Status: harness.DetectionAvailable, Available: true},
		harness.DetectionResult{Status: harness.DetectionNotInstalled},
		harness.DetectionResult{Status: harness.DetectionConfiguredUnverified, Configured: true},
		harness.DetectionResult{Status: harness.DetectionInstalledNoCanvas, Installed: true},
	)
	state := newInitIntegrationSelectorState(options)

	view := renderInitIntegrationSelectorView(state)

	if strings.Contains(strings.ReplaceAll(view, "\r\n", ""), "\n") {
		t.Fatalf("view contains bare newline; raw terminal render must use CRLF: %q", view)
	}
	for _, want := range []string{
		"\r\n> [X] Engram",
		"\r\n  [X] Stitch",
		"\r\n  [ ] OpenDesign",
		"\r\n  [ ] OpenPencil",
		"\r\n    memoria persistente y decisiones del proyecto",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q in %q", want, view)
		}
	}
}

func assertSelectedIntegrations(t *testing.T, got, want map[string]bool) {
	t.Helper()
	for _, key := range []string{"engram", "stitch", "opendesign", "openpencil"} {
		if got[key] != want[key] {
			t.Fatalf("selected[%s] = %v, want %v; full=%v", key, got[key], want[key], got)
		}
	}
}
