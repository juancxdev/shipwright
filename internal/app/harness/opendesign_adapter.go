package harness

import "fmt"

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
	return mustRenderTemplate("templates/project/harness/design/opendesign-task.md", RenderVars{
		"request": requestOrProjectName(state, request),
	})
}

func generateOpenDesignPrototypePlaceholder(state *State, request string) string {
	return mustRenderTemplate("templates/project/harness/design/opendesign-prototype.md", RenderVars{
		"request": requestOrProjectName(state, request),
	})
}

func generateOpenDesignResponsiveQAPlaceholder() string {
	return mustRenderTemplate("templates/project/harness/design/opendesign-responsive-qa.md", nil)
}
