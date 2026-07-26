package harness

import "fmt"

type OpenPencilDesignAdapter struct{}

func NewOpenPencilDesignAdapter() *OpenPencilDesignAdapter {
	return &OpenPencilDesignAdapter{}
}

func (o *OpenPencilDesignAdapter) AdapterName() string {
	return DesignModeOpenPencil
}

func (o *OpenPencilDesignAdapter) StartDesign(state *State, request string) (*DesignResult, error) {
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

	if err := ensureDir(DesignOpenPencilDir); err != nil {
		return nil, fmt.Errorf("cannot create openpencil dir: %w", err)
	}
	if err := ensureDir(DesignExportsDir); err != nil {
		return nil, fmt.Errorf("cannot create exports dir: %w", err)
	}

	task := generateDesignTask(state, request)
	if err := WriteFile(DesignTaskFile, task); err != nil {
		return nil, fmt.Errorf("cannot write design-task.md: %w", err)
	}
	files = append(files, DesignTaskFile)

	penFile := ".harness/artifacts/design/openpencil/app.pen"

	if err := SaveDesignState(DesignModeOpenPencil, false); err != nil {
		return nil, fmt.Errorf("cannot save design state: %w", err)
	}

	return &DesignResult{
		Adapter:      DesignModeOpenPencil,
		Mode:         DesignModeOpenPencil,
		FilesCreated: files,
		PenFile:      penFile,
		TaskFile:     DesignTaskFile,
		Message:      "OpenPencil design task created. AI agent should read .harness/artifacts/design/openpencil/design-task.md and use open-pencil_* MCP tools to create the .pen file.",
	}, nil
}

func (o *OpenPencilDesignAdapter) Status() (*DesignStatus, error) {
	return &DesignStatus{
		Adapter:              DesignModeOpenPencil,
		Mode:                 DesignModeOpenPencil,
		Available:            true,
		PenFile:              ".harness/artifacts/design/openpencil/app.pen",
		HasBrief:             ArtifactExists(".harness/artifacts/design/ux-brief.md"),
		HasFlows:             ArtifactExists(".harness/artifacts/design/user-flows.md"),
		HasDecisions:         ArtifactExists(".harness/artifacts/design/design-decisions.md"),
		HasPrototype:         ArtifactExists(".harness/artifacts/design/prototype.md"),
		HasWireframes:        ArtifactExists(".harness/artifacts/design/wireframes.md"),
		HasTaskFile:          ArtifactExists(DesignTaskFile),
		HasResponsiveQA:      ArtifactExists(".harness/artifacts/design/responsive-qa.md"),
		HasRouteInventory:    ArtifactExists(DesignRouteInventoryFile),
		HasAssetManifest:     ArtifactExists(DesignAssetManifestFile),
		HasSourceScreenshots: dirHasEvidenceFiles(DesignSourceScreenshotsDir),
		HasFidelityReport:    ArtifactExists(DesignFidelityReportFile),
		GateChecks:           EvaluateDesignEvidenceGates(nil),
	}, nil
}

func generateDesignTask(state *State, request string) string {
	return mustRenderTemplate("templates/project/harness/design/openpencil-task.md", RenderVars{
		"project_name": state.ProjectName,
		"request":      request,
	})
}

func ensureDir(path string) error {
	return WriteFile(path+"/.gitkeep", "")
}
