package harness

import "fmt"

type DocOnlyDesignFallback struct{}

func NewDocOnlyDesignFallback() *DocOnlyDesignFallback {
	return &DocOnlyDesignFallback{}
}

func (d *DocOnlyDesignFallback) AdapterName() string {
	return DesignModeDocOnly
}

func (d *DocOnlyDesignFallback) StartDesign(state *State, request string) (*DesignResult, error) {
	var files []string

	brief := generateUXBrief(state, request)
	if err := WriteFile(".harness/artifacts/design/ux-brief.md", brief); err != nil {
		return nil, fmt.Errorf("cannot write ux-brief.md: %w", err)
	}
	files = append(files, ".harness/artifacts/design/ux-brief.md")

	flows := generateUserFlows(state)
	if err := WriteFile(".harness/artifacts/design/user-flows.md", flows); err != nil {
		return nil, fmt.Errorf("cannot write user-flows.md: %w", err)
	}
	files = append(files, ".harness/artifacts/design/user-flows.md")

	decisions := generateDesignDecisions(state)
	if err := WriteFile(".harness/artifacts/design/design-decisions.md", decisions); err != nil {
		return nil, fmt.Errorf("cannot write design-decisions.md: %w", err)
	}
	files = append(files, ".harness/artifacts/design/design-decisions.md")

	wireframes := generateWireframesDoc(state)
	if err := WriteFile(".harness/artifacts/design/wireframes.md", wireframes); err != nil {
		return nil, fmt.Errorf("cannot write wireframes.md: %w", err)
	}
	files = append(files, ".harness/artifacts/design/wireframes.md")

	prototype := generatePrototypeDoc(state)
	if err := WriteFile(".harness/artifacts/design/prototype.md", prototype); err != nil {
		return nil, fmt.Errorf("cannot write prototype.md: %w", err)
	}
	files = append(files, ".harness/artifacts/design/prototype.md")

	responsiveQA := generateResponsiveQADoc(state)
	if err := WriteFile(".harness/artifacts/design/responsive-qa.md", responsiveQA); err != nil {
		return nil, fmt.Errorf("cannot write responsive-qa.md: %w", err)
	}
	files = append(files, ".harness/artifacts/design/responsive-qa.md")

	if err := SaveDesignState(DesignModeDocOnly, true); err != nil {
		return nil, fmt.Errorf("cannot save design state: %w", err)
	}

	return &DesignResult{
		Adapter:      DesignModeDocOnly,
		Mode:         DesignModeDocOnly,
		FilesCreated: files,
		Message:      "OpenPencil unavailable: design generated in doc-only mode.",
		FallbackUsed: true,
	}, nil
}

func (d *DocOnlyDesignFallback) Status() (*DesignStatus, error) {
	mode, _, _ := LoadDesignState()

	return &DesignStatus{
		Adapter:              DesignModeDocOnly,
		Mode:                 mode,
		Available:            false,
		HasBrief:             ArtifactExists(".harness/artifacts/design/ux-brief.md"),
		HasFlows:             ArtifactExists(".harness/artifacts/design/user-flows.md"),
		HasDecisions:         ArtifactExists(".harness/artifacts/design/design-decisions.md"),
		HasPrototype:         ArtifactExists(".harness/artifacts/design/prototype.md"),
		HasWireframes:        ArtifactExists(".harness/artifacts/design/wireframes.md"),
		HasTaskFile:          false,
		HasResponsiveQA:      ArtifactExists(".harness/artifacts/design/responsive-qa.md"),
		HasRouteInventory:    ArtifactExists(DesignRouteInventoryFile),
		HasAssetManifest:     ArtifactExists(DesignAssetManifestFile),
		HasSourceScreenshots: dirHasEvidenceFiles(DesignSourceScreenshotsDir),
		HasFidelityReport:    ArtifactExists(DesignFidelityReportFile),
		GateChecks:           EvaluateDesignEvidenceGates(nil),
	}, nil
}

func generateUXBrief(state *State, request string) string {
	return mustRenderTemplate("templates/project/harness/design/ux-brief.md", RenderVars{
		"project_name": state.ProjectName,
		"request":      request,
	})
}

func generateUserFlows(state *State) string {
	return mustRenderTemplate("templates/project/harness/design/user-flows.md", RenderVars{
		"project_name": state.ProjectName,
	})
}

func generateDesignDecisions(state *State) string {
	return mustRenderTemplate("templates/project/harness/design/design-decisions.md", RenderVars{
		"project_name": state.ProjectName,
	})
}

func generateWireframesDoc(state *State) string {
	return mustRenderTemplate("templates/project/harness/design/doconly-wireframes.md", RenderVars{
		"project_name": state.ProjectName,
	})
}

func generatePrototypeDoc(state *State) string {
	return mustRenderTemplate("templates/project/harness/design/doconly-prototype.md", RenderVars{
		"project_name": state.ProjectName,
	})
}

func generateResponsiveQADoc(state *State) string {
	return mustRenderTemplate("templates/project/harness/design/doconly-responsive-qa.md", RenderVars{
		"project_name": state.ProjectName,
	})
}
