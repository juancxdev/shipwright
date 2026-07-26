package harness

import (
	"os"
	"path/filepath"
	"strings"

	designapp "shipwright/internal/design/application"
	designdomain "shipwright/internal/design/domain"
)

const (
	GateBaselineCaptured  = designdomain.GateBaselineCaptured
	GateAssetsPreserved   = designdomain.GateAssetsPreserved
	GateProviderPublished = designdomain.GateProviderPublished
	GateFidelityVerified  = designdomain.GateFidelityVerified
	GateTokenQuotaOK      = designdomain.GateTokenQuotaOK
)

const (
	DesignProviderFallbackApprovalFile = ".harness/approvals/design-provider-fallback.json"
	DesignAssetChangeApprovalFile      = ".harness/approvals/design-asset-change.json"
)

type DesignGateCheck = designdomain.GateCheck

func EvaluateDesignEvidenceGates(state *State) []DesignGateCheck {
	mode, _, _ := LoadDesignState()
	if mode == "" {
		mode = DesignModeDocOnly
	}
	baselineRequired := DesignBaselineRequired()
	checks := []DesignGateCheck{
		evaluateBaselineCaptured(baselineRequired),
		evaluateAssetsPreserved(baselineRequired),
		evaluateProviderPublished(mode),
		evaluateFidelityVerified(baselineRequired),
		evaluateTokenQuotaOK(mode),
	}
	return checks
}

func DesignBaselineRequired() bool {
	profile, err := LoadProjectProfile()
	if err != nil || profile == nil {
		return false
	}
	return profile.ExistingProject && profile.Structure.HasFrontend
}

func DesignApprovalMissingArtifacts(state *State) []string {
	return designapp.MissingBlockingChecks(EvaluateDesignEvidenceGates(state))
}

func FindDesignGateCheck(checks []DesignGateCheck, gate string) *DesignGateCheck {
	for i := range checks {
		if checks[i].Gate == gate {
			return &checks[i]
		}
	}
	return nil
}

func RenderDesignGateSummary(checks []DesignGateCheck) string {
	return designapp.RenderGateSummary(checks)
}

func evaluateBaselineCaptured(required bool) DesignGateCheck {
	if !required {
		return DesignGateCheck{Gate: GateBaselineCaptured, Pass: true, Blocking: false, Message: "baseline not required for greenfield/non-frontend project"}
	}
	if !ArtifactExists(DesignRouteInventoryFile) {
		return DesignGateCheck{Gate: GateBaselineCaptured, Pass: false, Blocking: true, Message: DesignRouteInventoryFile + " is missing"}
	}
	if !dirHasEvidenceFiles(DesignSourceScreenshotsDir) {
		return DesignGateCheck{Gate: GateBaselineCaptured, Pass: false, Blocking: true, Message: DesignSourceScreenshotsDir + " has no source screenshot evidence"}
	}
	return DesignGateCheck{Gate: GateBaselineCaptured, Pass: true, Blocking: true, Message: "route inventory and source screenshots found"}
}

func evaluateAssetsPreserved(required bool) DesignGateCheck {
	if !required {
		return DesignGateCheck{Gate: GateAssetsPreserved, Pass: true, Blocking: false, Message: "asset preservation baseline not required"}
	}
	manifest, err := LoadDesignAssetManifest()
	if err != nil {
		return DesignGateCheck{Gate: GateAssetsPreserved, Pass: false, Blocking: true, Message: DesignAssetManifestFile + " is missing or invalid"}
	}
	if len(manifest.Assets) == 0 {
		return DesignGateCheck{Gate: GateAssetsPreserved, Pass: false, Blocking: true, Message: "asset manifest has no assets"}
	}
	if !AssetManifestHasLogo(manifest) && !ArtifactExists(DesignAssetChangeApprovalFile) {
		return DesignGateCheck{Gate: GateAssetsPreserved, Pass: false, Blocking: true, Message: "asset manifest has no logo; approve asset changes explicitly if the project truly has no logo"}
	}
	if report := readOptionalArtifact(DesignFidelityReportFile); strings.Contains(strings.ToLower(report), "asset mismatch") || strings.Contains(strings.ToLower(report), "logo mismatch") || strings.Contains(strings.ToLower(report), "logo changed") {
		if !ArtifactExists(DesignAssetChangeApprovalFile) {
			return DesignGateCheck{Gate: GateAssetsPreserved, Pass: false, Blocking: true, Message: "fidelity report mentions asset/logo mismatch without explicit design-asset-change approval"}
		}
	}
	return DesignGateCheck{Gate: GateAssetsPreserved, Pass: true, Blocking: true, Message: "asset manifest is present and no unapproved asset mismatch was found"}
}

func evaluateProviderPublished(mode string) DesignGateCheck {
	if mode == DesignModeDocOnly {
		if ArtifactExists(DesignProviderFallbackApprovalFile) {
			return DesignGateCheck{Gate: GateProviderPublished, Pass: true, Blocking: true, Message: "doc-only fallback explicitly accepted"}
		}
		if DesignBaselineRequired() {
			return DesignGateCheck{Gate: GateProviderPublished, Pass: false, Blocking: true, Message: "existing frontend requires provider publish evidence or explicit doc-only fallback approval"}
		}
		return DesignGateCheck{Gate: GateProviderPublished, Pass: true, Blocking: false, Message: "doc-only provider does not publish visual artifacts"}
	}
	reportPath := providerReportPath(mode)
	if reportPath == "" || !ArtifactExists(reportPath) {
		return DesignGateCheck{Gate: GateProviderPublished, Pass: false, Blocking: true, Message: "provider publish report is missing for " + mode}
	}
	report := strings.ToLower(readOptionalArtifact(reportPath))
	if strings.Contains(report, "quota") || strings.Contains(report, "token exhausted") || strings.Contains(report, "no tengo tokens") || strings.Contains(report, "insufficient tokens") {
		return DesignGateCheck{Gate: GateProviderPublished, Pass: false, Blocking: true, Message: "provider report indicates token/quota exhaustion"}
	}
	if strings.Contains(report, "status: pass") || strings.Contains(report, "status: complete") || strings.Contains(report, "status: published") || strings.Contains(report, "create_artifact succeeded") || strings.Contains(report, "publish succeeded") {
		return DesignGateCheck{Gate: GateProviderPublished, Pass: true, Blocking: true, Message: "provider publish evidence accepted"}
	}
	if strings.Contains(report, "status: partial") || strings.Contains(report, "partial") || strings.Contains(report, "publish failed") || strings.Contains(report, "import failed") {
		if ArtifactExists(DesignProviderFallbackApprovalFile) {
			return DesignGateCheck{Gate: GateProviderPublished, Pass: true, Blocking: true, Message: "partial provider result explicitly accepted"}
		}
		return DesignGateCheck{Gate: GateProviderPublished, Pass: false, Blocking: true, Message: "provider report is partial/failed without explicit fallback approval"}
	}
	return DesignGateCheck{Gate: GateProviderPublished, Pass: false, Blocking: true, Message: "provider report does not prove publish success"}
}

func evaluateFidelityVerified(required bool) DesignGateCheck {
	if !required && !ArtifactExists(DesignFidelityReportFile) {
		return DesignGateCheck{Gate: GateFidelityVerified, Pass: true, Blocking: false, Message: "fidelity report not required for greenfield/non-frontend project"}
	}
	if !ArtifactExists(DesignFidelityReportFile) {
		return DesignGateCheck{Gate: GateFidelityVerified, Pass: false, Blocking: true, Message: DesignFidelityReportFile + " is missing"}
	}
	report := strings.ToLower(readOptionalArtifact(DesignFidelityReportFile))
	if strings.Contains(report, "status: fail") || strings.Contains(report, "status: failed") {
		return DesignGateCheck{Gate: GateFidelityVerified, Pass: false, Blocking: true, Message: "fidelity report status is fail"}
	}
	if strings.Contains(report, "status: partial") || strings.Contains(report, "conditional-pass") || strings.Contains(report, "approximation-from-code") {
		if ArtifactExists(DesignProviderFallbackApprovalFile) {
			return DesignGateCheck{Gate: GateFidelityVerified, Pass: true, Blocking: true, Message: "conditional fidelity explicitly accepted"}
		}
		return DesignGateCheck{Gate: GateFidelityVerified, Pass: false, Blocking: true, Message: "fidelity is conditional/partial without explicit acceptance"}
	}
	if strings.Contains(report, "status: pass") || strings.Contains(report, "fidelity: pass") {
		return DesignGateCheck{Gate: GateFidelityVerified, Pass: true, Blocking: true, Message: "fidelity report passed"}
	}
	return DesignGateCheck{Gate: GateFidelityVerified, Pass: false, Blocking: true, Message: "fidelity report does not include Status: pass"}
}

func evaluateTokenQuotaOK(mode string) DesignGateCheck {
	paths := []string{DesignFidelityReportFile, providerReportPath(mode)}
	for _, path := range paths {
		if path == "" || !ArtifactExists(path) {
			continue
		}
		content := strings.ToLower(readOptionalArtifact(path))
		if strings.Contains(content, "quota") || strings.Contains(content, "token exhausted") || strings.Contains(content, "insufficient tokens") || strings.Contains(content, "no tengo tokens") {
			return DesignGateCheck{Gate: GateTokenQuotaOK, Pass: false, Blocking: true, Message: "design evidence reports token/quota exhaustion"}
		}
	}
	return DesignGateCheck{Gate: GateTokenQuotaOK, Pass: true, Blocking: true, Message: "no token/quota blocker found in design evidence"}
}

func providerReportPath(mode string) string {
	switch mode {
	case DesignModeStitch:
		return ".harness/artifacts/design/stitch/stitch-report.md"
	case DesignModeOpenDesign:
		return ".harness/artifacts/design/opendesign/opendesign-report.md"
	case DesignModeOpenPencil:
		return ".harness/artifacts/design/openpencil/save-status.md"
	default:
		return ""
	}
}

func dirHasEvidenceFiles(dir string) bool {
	resolved := ArtifactPath(dir)
	entries, err := os.ReadDir(resolved)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		return true
	}
	return false
}

func readOptionalArtifact(path string) string {
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(ArtifactPath(path))
	if err != nil {
		return ""
	}
	return string(data)
}

func WriteDesignGateReport(state *State) error {
	checks := EvaluateDesignEvidenceGates(state)
	content := mustRenderTemplate("templates/project/harness/runtime/design-gates.md", RenderVars{
		"gate_summary": RenderDesignGateSummary(checks),
	})
	return WriteFile(".harness/artifacts/design/design-gates.md", content)
}

func EnsureDesignEvidenceDirs() error {
	for _, dir := range []string{DesignBaselineDir, DesignSourceScreenshotsDir, DesignStitchDir, DesignOpenDesignDir, DesignOpenPencilDir} {
		if err := os.MkdirAll(ArtifactPath(filepath.ToSlash(dir)), 0755); err != nil {
			return err
		}
	}
	return nil
}
