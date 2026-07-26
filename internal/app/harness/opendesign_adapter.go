package harness

import (
	"fmt"
	"strings"
)

type OpenDesignAdapter struct{}

func NewOpenDesignAdapter() *OpenDesignAdapter { return &OpenDesignAdapter{} }

func (o *OpenDesignAdapter) AdapterName() string { return DesignModeOpenDesign }

func (o *OpenDesignAdapter) StartDesign(state *State, request string) (*DesignResult, error) {
	var files []string

	for _, dir := range []string{DesignDir, DesignOpenDesignDir} {
		if err := ensureDir(dir); err != nil {
			return nil, fmt.Errorf("cannot create %s: %w", dir, err)
		}
	}

	brief := generateUXBrief(state, request)
	if err := WriteFile(".harness/artifacts/design/ux-brief.md", brief); err != nil {
		return nil, err
	}
	files = append(files, ".harness/artifacts/design/ux-brief.md")

	flows := generateUserFlows(state)
	if err := WriteFile(".harness/artifacts/design/user-flows.md", flows); err != nil {
		return nil, err
	}
	files = append(files, ".harness/artifacts/design/user-flows.md")

	decisions := generateDesignDecisions(state)
	if err := WriteFile(".harness/artifacts/design/design-decisions.md", decisions); err != nil {
		return nil, err
	}
	files = append(files, ".harness/artifacts/design/design-decisions.md")

	task := generateOpenDesignTask(state, request)
	if err := WriteFile(DesignOpenDesignTaskFile, task); err != nil {
		return nil, err
	}
	files = append(files, DesignOpenDesignTaskFile)

	prototype := generateOpenDesignPrototypePlaceholder(state, request)
	if err := WriteFile(".harness/artifacts/design/prototype.md", prototype); err != nil {
		return nil, err
	}
	files = append(files, ".harness/artifacts/design/prototype.md")

	responsiveQA := generateOpenDesignResponsiveQAPlaceholder()
	if err := WriteFile(".harness/artifacts/design/responsive-qa.md", responsiveQA); err != nil {
		return nil, err
	}
	files = append(files, ".harness/artifacts/design/responsive-qa.md")

	if err := SaveDesignState(DesignModeOpenDesign, false); err != nil {
		return nil, err
	}

	return &DesignResult{
		Adapter:      DesignModeOpenDesign,
		Mode:         DesignModeOpenDesign,
		FilesCreated: files,
		TaskFile:     DesignOpenDesignTaskFile,
		Message:      "OpenDesign artifact task created. AI agent should use open-design MCP tools to create HTML/React design artifacts plus .artifact.json manifests and evidence under .harness/artifacts/design/opendesign/.",
	}, nil
}

func (o *OpenDesignAdapter) Status() (*DesignStatus, error) {
	return &DesignStatus{
		Adapter:              DesignModeOpenDesign,
		Mode:                 DesignModeOpenDesign,
		Available:            true,
		HasBrief:             ArtifactExists(".harness/artifacts/design/ux-brief.md"),
		HasFlows:             ArtifactExists(".harness/artifacts/design/user-flows.md"),
		HasDecisions:         ArtifactExists(".harness/artifacts/design/design-decisions.md"),
		HasPrototype:         ArtifactExists(".harness/artifacts/design/prototype.md"),
		HasResponsiveQA:      ArtifactExists(".harness/artifacts/design/responsive-qa.md"),
		HasRouteInventory:    ArtifactExists(DesignRouteInventoryFile),
		HasAssetManifest:     ArtifactExists(DesignAssetManifestFile),
		HasSourceScreenshots: dirHasEvidenceFiles(DesignSourceScreenshotsDir),
		HasFidelityReport:    ArtifactExists(DesignFidelityReportFile),
		GateChecks:           EvaluateDesignEvidenceGates(nil),
		HasTaskFile:          ArtifactExists(DesignOpenDesignTaskFile),
	}, nil
}

func generateOpenDesignTask(state *State, request string) string {
	var sb strings.Builder
	sb.WriteString("# OpenDesign Artifact Task\n\n")
	sb.WriteString("Use OpenDesign as an artifact design provider. Do not treat OpenDesign as OpenPencil/Figma canvas; it exposes project/file/artifact tools.\n\n")
	sb.WriteString("## Request\n\n")
	if request != "" {
		sb.WriteString(request + "\n\n")
	} else {
		sb.WriteString(state.ProjectName + "\n\n")
	}
	sb.WriteString("## Required skills\n\n")
	sb.WriteString("Load and follow `.opencode/skills/opendesign-generate-artifact/SKILL.md`. For existing UI baselines, also follow `existing-web-to-openpencil` for route/screenshot/fidelity discipline, but publish through OpenDesign artifacts rather than OpenPencil frames.\n\n")
	sb.WriteString("## MCP tools\n\n")
	sb.WriteString("Try these exact OpenCode tool names first: `open-design_get_active_context`, `open-design_list_projects`, `open-design_list_files`, `open-design_create_artifact`. If absent, check `opendesign_*` and `open_design_*`.\n\n")
	sb.WriteString("## Required outputs\n\n")
	sb.WriteString("- `.harness/artifacts/design/opendesign/<entry>.html` or equivalent artifact entry.\n")
	sb.WriteString("- `.harness/artifacts/design/opendesign/<entry>.html.artifact.json` sidecar manifest.\n")
	sb.WriteString("- `.harness/artifacts/design/opendesign/opendesign-report.md` with MCP calls, active project/context, artifact ID/name, manifest path, and limitations.\n")
	sb.WriteString("- `.harness/artifacts/design/prototype.md`, `.harness/artifacts/design/responsive-qa.md`, and `.harness/artifacts/design/fidelity-report.md` for existing UI baselines.\n\n")
	sb.WriteString("## Guardrails\n\n")
	sb.WriteString("- If `ARTIFACT_MANIFEST_REQUIRED` appears, create `<entry>.artifact.json` and retry; do not mark pass without a manifest.\n")
	sb.WriteString("- If OpenDesign MCP publish/import fails, status is `partial` unless the user explicitly accepts local artifact fallback.\n")
	sb.WriteString("- Do not claim OpenDesign canvas/frame completion; OpenDesign completion means artifact creation/publish plus evidence.\n")
	return sb.String()
}

func generateOpenDesignPrototypePlaceholder(state *State, request string) string {
	var sb strings.Builder
	sb.WriteString("# Prototype\n\n")
	sb.WriteString("> OpenDesign artifact placeholder. Replace with OpenDesign artifact links, local artifact paths, manifest paths, and interaction notes.\n\n")
	sb.WriteString("## Provider\n\nOpenDesign\n\n")
	sb.WriteString("## Request\n\n")
	if request != "" {
		sb.WriteString(request + "\n")
	} else {
		sb.WriteString(state.ProjectName + "\n")
	}
	return sb.String()
}

func generateOpenDesignResponsiveQAPlaceholder() string {
	return `# Responsive & Accessibility QA

> OpenDesign artifact placeholder. Replace after the generated artifact is inspected at mobile/tablet/desktop widths.

## Required OpenDesign evidence

| Route/View | Mobile | Tablet | Desktop | Notes |
|------------|--------|--------|---------|-------|
| TBD | Pending | Pending | Pending | Generate artifact, manifest, and inspect in browser/OpenDesign. |

## Gates

- [ ] OpenDesign artifact generated with sidecar .artifact.json manifest.
- [ ] Artifact opened in browser/OpenDesign.
- [ ] Mobile/tablet/desktop responsive states inspected.
- [ ] No clipped content, broken layout, or accidental horizontal scrolling.
- [ ] Touch targets are >= 44x44 on mobile.
- [ ] Existing UI baseline, if any, was compared against source screenshots.
`
}
