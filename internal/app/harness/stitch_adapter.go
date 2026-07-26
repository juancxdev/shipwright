package harness

import "fmt"

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
		Adapter:              DesignModeStitch,
		Mode:                 DesignModeStitch,
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
		HasTaskFile:          ArtifactExists(DesignStitchTaskFile),
	}, nil
}

func generateStitchDesignMD(state *State, request string) string {
	requestLine := ""
	if request != "" {
		requestLine = "- Request: " + request + "\n"
	}
	return mustRenderTemplate("templates/project/harness/design/stitch-design.md", RenderVars{
		"project_name": state.ProjectName,
		"request_line": requestLine,
	})
}

func generateStitchDesignTask(state *State, request string) string {
	return mustRenderTemplate("templates/project/harness/design/stitch-task.md", RenderVars{
		"request": requestOrProjectName(state, request),
	})
}

func generateStitchPrototypePlaceholder(state *State, request string) string {
	return mustRenderTemplate("templates/project/harness/design/stitch-prototype.md", RenderVars{
		"request": requestOrProjectName(state, request),
	})
}

func generateStitchResponsiveQAPlaceholder() string {
	return mustRenderTemplate("templates/project/harness/design/stitch-responsive-qa.md", nil)
}
