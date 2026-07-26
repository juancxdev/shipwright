package harness

import "path/filepath"

type adapterDesignProvider struct {
	name    string
	adapter DesignPort
}

func NewOpenDesignProvider() DesignProvider {
	return &adapterDesignProvider{name: DesignModeOpenDesign, adapter: NewOpenDesignAdapter()}
}

func NewStitchProvider() DesignProvider {
	return &adapterDesignProvider{name: DesignModeStitch, adapter: NewStitchDesignAdapter()}
}

func NewOpenPencilProvider() DesignProvider {
	return &adapterDesignProvider{name: DesignModeOpenPencil, adapter: NewOpenPencilDesignAdapter()}
}

func NewDocOnlyProvider() DesignProvider {
	return &adapterDesignProvider{name: DesignModeDocOnly, adapter: NewDocOnlyDesignFallback()}
}

func (p *adapterDesignProvider) Name() string { return p.name }

func (p *adapterDesignProvider) Detect() DesignProviderDetection {
	status, err := p.adapter.Status()
	if err != nil {
		return DesignProviderDetection{Name: p.name, Available: false, Reason: err.Error()}
	}
	return DesignProviderDetection{Name: p.name, Available: status.Available, Mode: status.Mode}
}

func (p *adapterDesignProvider) Prepare(state *State, request string) (*DesignProviderResult, error) {
	return &DesignProviderResult{Provider: p.name, Status: DesignProviderStatusReady, Message: "provider preparation delegated to Generate"}, nil
}

func (p *adapterDesignProvider) Generate(state *State, request string) (*DesignProviderResult, error) {
	result, err := p.adapter.StartDesign(state, request)
	if err != nil {
		return nil, err
	}
	return &DesignProviderResult{Provider: p.name, Status: DesignProviderStatusReady, Files: result.FilesCreated, Message: result.Message}, nil
}

func (p *adapterDesignProvider) Publish(state *State) (*DesignProviderResult, error) {
	checks := EvaluateDesignEvidenceGates(state)
	if check := FindDesignGateCheck(checks, GateProviderPublished); check != nil && !check.Pass {
		return &DesignProviderResult{Provider: p.name, Status: DesignProviderStatusPartial, Message: check.Message}, nil
	}
	return &DesignProviderResult{Provider: p.name, Status: DesignProviderStatusReady, Message: "provider publish evidence accepted"}, nil
}

func (p *adapterDesignProvider) Verify(state *State) (*DesignProviderResult, error) {
	checks := EvaluateDesignEvidenceGates(state)
	status := DesignProviderStatusReady
	for _, check := range checks {
		if check.Blocking && !check.Pass {
			status = DesignProviderStatusBlocked
			break
		}
	}
	return &DesignProviderResult{Provider: p.name, Status: status, Message: RenderDesignGateSummary(checks)}, nil
}

func (p *adapterDesignProvider) Report(state *State) (*DesignProviderResult, error) {
	result, err := p.Verify(state)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(DesignDir, p.name, p.name+"-provider-report.md")
	if p.name == DesignModeDocOnly {
		path = filepath.Join(DesignDir, "doc-only-provider-report.md")
	}
	if err := writeDesignProviderReport(path, p.name, result.Status, result.Message, result.Files); err != nil {
		return nil, err
	}
	result.Files = append(result.Files, path)
	return result, nil
}
