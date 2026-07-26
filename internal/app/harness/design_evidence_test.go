package harness

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestUXApprovalBlocksMissingExistingFrontendBaselineEvidence(t *testing.T) {
	chdirTemp(t)
	profile := &ProjectProfile{ProjectName: "astro", ExistingProject: true, Structure: ProjectStructure{HasFrontend: true}}
	if err := SaveProjectProfile(profile); err != nil {
		t.Fatalf("SaveProjectProfile: %v", err)
	}
	state := NewState("astro")
	state.CurrentPhase = StateUXApproval

	missing := DesignApprovalMissingArtifacts(state)
	joined := strings.Join(missing, "\n")
	for _, want := range []string{GateBaselineCaptured, GateAssetsPreserved, GateProviderPublished, GateFidelityVerified} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing UX gate %s in:\n%s", want, joined)
		}
	}
}

func TestUXApprovalPassesEvidenceForExistingFrontend(t *testing.T) {
	chdirTemp(t)
	profile := &ProjectProfile{ProjectName: "astro", ExistingProject: true, Structure: ProjectStructure{HasFrontend: true}}
	if err := SaveProjectProfile(profile); err != nil {
		t.Fatalf("SaveProjectProfile: %v", err)
	}
	writeTestFile(t, DesignRouteInventoryFile, "# Routes\n\n- /\n")
	writeTestFile(t, filepath.Join(DesignSourceScreenshotsDir, "home-desktop.png"), "fake png")
	writeTestFile(t, DesignAssetManifestFile, `{"version":"1","assets":[{"path":"public/logo.svg","role":"logo"}]}`)
	writeTestFile(t, DesignFidelityReportFile, "# Fidelity\n\nStatus: pass\n")
	writeTestFile(t, ".harness/artifacts/design/stitch/stitch-report.md", "# Stitch\n\nStatus: published\n")
	if err := SaveDesignState(DesignModeStitch, false); err != nil {
		t.Fatalf("SaveDesignState: %v", err)
	}
	state := NewState("astro")
	state.CurrentPhase = StateUXApproval

	if missing := DesignApprovalMissingArtifacts(state); len(missing) != 0 {
		t.Fatalf("unexpected missing UX evidence: %+v", missing)
	}
}

func TestUXApprovalBlocksLogoMismatchWithoutExplicitApproval(t *testing.T) {
	chdirTemp(t)
	profile := &ProjectProfile{ProjectName: "astro", ExistingProject: true, Structure: ProjectStructure{HasFrontend: true}}
	_ = SaveProjectProfile(profile)
	writeTestFile(t, DesignRouteInventoryFile, "# Routes\n\n- /\n")
	writeTestFile(t, filepath.Join(DesignSourceScreenshotsDir, "home-desktop.png"), "fake png")
	writeTestFile(t, DesignAssetManifestFile, `{"version":"1","assets":[{"path":"public/logo.svg","role":"logo"}]}`)
	writeTestFile(t, DesignFidelityReportFile, "# Fidelity\n\nStatus: pass\n\nlogo mismatch\n")
	writeTestFile(t, ".harness/artifacts/design/stitch/stitch-report.md", "# Stitch\n\nStatus: published\n")
	_ = SaveDesignState(DesignModeStitch, false)

	missing := strings.Join(DesignApprovalMissingArtifacts(NewState("astro")), "\n")
	if !strings.Contains(missing, GateAssetsPreserved) {
		t.Fatalf("expected assets-preserved blocker, got:\n%s", missing)
	}
	if err := WriteFile(DesignAssetChangeApprovalFile, "{}"); err != nil {
		t.Fatalf("write approval: %v", err)
	}
	missing = strings.Join(DesignApprovalMissingArtifacts(NewState("astro")), "\n")
	if strings.Contains(missing, GateAssetsPreserved) {
		t.Fatalf("asset approval should unblock assets-preserved, got:\n%s", missing)
	}
}
