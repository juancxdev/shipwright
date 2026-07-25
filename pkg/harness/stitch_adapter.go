package harness

import (
	"fmt"
	"strings"
)

type StitchDesignAdapter struct{}

func NewStitchDesignAdapter() *StitchDesignAdapter {
	return &StitchDesignAdapter{}
}

func (s *StitchDesignAdapter) AdapterName() string {
	return DesignModeStitch
}

func (s *StitchDesignAdapter) StartDesign(state *State, request string) (*DesignResult, error) {
	var files []string

	if err := ensureDir(DesignDir); err != nil {
		return nil, fmt.Errorf("cannot create design dir: %w", err)
	}
	if err := ensureDir(DesignStitchDir); err != nil {
		return nil, fmt.Errorf("cannot create stitch dir: %w", err)
	}
	if err := ensureDir(DesignStitchScreensDir); err != nil {
		return nil, fmt.Errorf("cannot create stitch screens dir: %w", err)
	}
	if err := ensureDir(DesignStitchExportsDir); err != nil {
		return nil, fmt.Errorf("cannot create stitch exports dir: %w", err)
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

	designMD := generateStitchDesignMD(state, request)
	if err := WriteFile(DesignStitchDesignMDFile, designMD); err != nil {
		return nil, err
	}
	files = append(files, DesignStitchDesignMDFile)

	task := generateStitchDesignTask(state, request)
	if err := WriteFile(DesignStitchTaskFile, task); err != nil {
		return nil, err
	}
	files = append(files, DesignStitchTaskFile)

	prototype := generateStitchPrototypePlaceholder(state, request)
	if err := WriteFile(".harness/artifacts/design/prototype.md", prototype); err != nil {
		return nil, err
	}
	files = append(files, ".harness/artifacts/design/prototype.md")

	responsiveQA := generateStitchResponsiveQAPlaceholder()
	if err := WriteFile(".harness/artifacts/design/responsive-qa.md", responsiveQA); err != nil {
		return nil, err
	}
	files = append(files, ".harness/artifacts/design/responsive-qa.md")

	if err := SaveDesignState(DesignModeStitch, false); err != nil {
		return nil, err
	}

	return &DesignResult{
		Adapter:      DesignModeStitch,
		Mode:         DesignModeStitch,
		FilesCreated: files,
		TaskFile:     DesignStitchTaskFile,
		Message:      "Stitch design task created. AI agent should use Google Stitch SDK/MCP to generate high-fidelity screens, screenshots, HTML exports, and evidence under .harness/artifacts/design/stitch/.",
	}, nil
}

func (s *StitchDesignAdapter) Status() (*DesignStatus, error) {
	return &DesignStatus{
		Adapter:         DesignModeStitch,
		Mode:            DesignModeStitch,
		Available:       true,
		HasBrief:        ArtifactExists(".harness/artifacts/design/ux-brief.md"),
		HasFlows:        ArtifactExists(".harness/artifacts/design/user-flows.md"),
		HasDecisions:    ArtifactExists(".harness/artifacts/design/design-decisions.md"),
		HasPrototype:    ArtifactExists(".harness/artifacts/design/prototype.md"),
		HasResponsiveQA: ArtifactExists(".harness/artifacts/design/responsive-qa.md"),
		HasTaskFile:     ArtifactExists(DesignStitchTaskFile),
	}, nil
}

func generateStitchDesignMD(state *State, request string) string {
	var sb strings.Builder
	sb.WriteString("# DESIGN.md\n\n")
	sb.WriteString("> Design system and generation rules for Google Stitch.\n\n")
	sb.WriteString("## Project\n\n")
	sb.WriteString("- Name: " + state.ProjectName + "\n")
	if request != "" {
		sb.WriteString("- Request: " + request + "\n")
	}
	sb.WriteString("\n## Design principles\n\n")
	sb.WriteString("- Generate high-fidelity UI, not low-fidelity wireframes.\n")
	sb.WriteString("- Preserve product scope and user goals from Shipwright artifacts.\n")
	sb.WriteString("- Create responsive mobile, tablet, and desktop variants.\n")
	sb.WriteString("- Prefer reusable components, clear hierarchy, accessible contrast, and realistic copy.\n")
	sb.WriteString("- Export screenshots and HTML evidence for every generated screen.\n")
	sb.WriteString("\n## Required outputs\n\n")
	sb.WriteString("- Stitch project/screen IDs.\n")
	sb.WriteString("- Screenshot exports.\n")
	sb.WriteString("- HTML exports.\n")
	sb.WriteString("- Design-to-code component map when frontend components exist.\n")
	return sb.String()
}

func generateStitchDesignTask(state *State, request string) string {
	var sb strings.Builder
	sb.WriteString("# Stitch Design Task\n\n")
	sb.WriteString("Use Google Stitch as the primary design provider for this project. Do not use OpenPencil unless the user explicitly asks for it.\n\n")
	sb.WriteString("## Request\n\n")
	if request != "" {
		sb.WriteString(request + "\n\n")
	} else {
		sb.WriteString(state.ProjectName + "\n\n")
	}
	sb.WriteString("## Credentials\n\n")
	sb.WriteString("Use Stitch only when `STITCH_API_KEY` is set, or `STITCH_ACCESS_TOKEN` + `GOOGLE_CLOUD_PROJECT` are set. If credentials are unavailable, report blocked and continue with doc-only artifacts only after explaining the limitation.\n\n")
	sb.WriteString("## Workflow\n\n")
	sb.WriteString("1. Read `.harness/artifacts/product/context.md`, `.harness/artifacts/product/scope.md`, and `.harness/artifacts/design/stitch/DESIGN.md`.\n")
	sb.WriteString("2. If recreating an existing UI, capture source route screenshots first and write `.harness/artifacts/design/route-inventory.md`.\n")
	sb.WriteString("3. Use Stitch SDK/MCP to create or update a Stitch project and generate mobile/tablet/desktop screens.\n")
	sb.WriteString("4. Generate variants only when useful; choose one recommended direction and document alternatives.\n")
	sb.WriteString("5. Export screenshots to `.harness/artifacts/design/stitch/exports/`.\n")
	sb.WriteString("6. Export HTML to `.harness/artifacts/design/stitch/html/` when available.\n")
	sb.WriteString("7. Write `.harness/artifacts/design/stitch/stitch-report.md` with project ID, screen IDs, prompts, exports, and known limitations.\n")
	sb.WriteString("8. Write or update `.harness/artifacts/design/prototype.md`, `.harness/artifacts/design/responsive-qa.md`, `.harness/artifacts/design/design-decisions.md`, and `.harness/artifacts/design/code-component-map.md` when components exist.\n")
	sb.WriteString("9. For existing UI baselines, compare Stitch screenshots against source screenshots in `.harness/artifacts/design/fidelity-report.md`; do not pass if the design materially diverges.\n\n")
	sb.WriteString("## Guardrails\n\n")
	sb.WriteString("- Stitch is the design provider; OpenPencil is not required.\n")
	sb.WriteString("- Do not claim fidelity without source screenshot vs Stitch screenshot comparison.\n")
	sb.WriteString("- Do not claim implementation readiness without code-component-map for reusable UI.\n")
	sb.WriteString("- Do not commit generated HTML as production frontend unless the frontend engineer explicitly accepts it as source.\n")
	return sb.String()
}

func generateStitchPrototypePlaceholder(state *State, request string) string {
	var sb strings.Builder
	sb.WriteString("# Prototype\n\n")
	sb.WriteString("> Stitch-first design placeholder. The UI/UX Designer must replace this with Stitch screen IDs, screenshots, HTML export references, and interaction notes.\n\n")
	sb.WriteString("## Provider\n\nGoogle Stitch\n\n")
	sb.WriteString("## Request\n\n")
	if request != "" {
		sb.WriteString(request + "\n")
	} else {
		sb.WriteString(state.ProjectName + "\n")
	}
	return sb.String()
}

func generateStitchResponsiveQAPlaceholder() string {
	return `# Responsive & Accessibility QA

> Stitch-first placeholder. Replace after generated screenshots are exported and inspected.

## Required Stitch evidence

| Screen | Mobile | Tablet | Desktop | Notes |
|--------|--------|--------|---------|-------|
| TBD | Pending | Pending | Pending | Generate and inspect Stitch exports. |

## Gates

- [ ] Stitch screenshots exported for mobile, tablet, and desktop.
- [ ] No clipped content or broken layout in exports.
- [ ] Touch targets are >= 44x44 on mobile.
- [ ] Contrast is acceptable for text and primary UI components.
- [ ] Existing UI baseline, if any, was compared against source screenshots.
`
}
