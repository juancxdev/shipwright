package harness

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	DesignModeOpenPencil = "openpencil"
	DesignModeDocOnly    = "doc-only"
)

const (
	DesignDir           = "design"
	DesignOpenPencilDir = "design/openpencil"
	DesignExportsDir    = "design/openpencil/exports"
	DesignTaskFile      = "design/openpencil/design-task.md"
	DesignStateFile     = ".harness/design-state.json"
)

type DesignResult struct {
	Adapter      string
	Mode         string
	FilesCreated []string
	PenFile      string
	TaskFile     string
	Message      string
	FallbackUsed bool
}

type DesignStatus struct {
	Adapter         string
	Mode            string
	Available       bool
	PenFile         string
	HasBrief        bool
	HasFlows        bool
	HasDecisions    bool
	HasPrototype    bool
	HasWireframes   bool
	HasTaskFile     bool
	HasResponsiveQA bool
}

type DesignPort interface {
	StartDesign(state *State, request string) (*DesignResult, error)
	Status() (*DesignStatus, error)
	AdapterName() string
}

type DesignService struct {
	primary      DesignPort
	fallback     DesignPort
	openpencilOn bool
}

func NewDesignService(integrations *Integrations) *DesignService {
	opOn := integrations != nil && integrations.OpenPencil.Enabled

	svc := &DesignService{
		fallback:     NewDocOnlyDesignFallback(),
		openpencilOn: opOn,
	}

	if opOn {
		svc.primary = NewOpenPencilDesignAdapter()
	} else {
		svc.primary = svc.fallback
	}

	return svc
}

func (ds *DesignService) StartDesign(state *State, request string) (*DesignResult, error) {
	if ds.openpencilOn {
		result, err := ds.primary.StartDesign(state, request)
		if err != nil {
			result, fbErr := ds.fallback.StartDesign(state, request)
			if fbErr != nil {
				return nil, fmt.Errorf("openpencil failed: %w; fallback also failed: %v", err, fbErr)
			}
			result.FallbackUsed = true
			result.Message = fmt.Sprintf("OpenPencil unavailable: design generated in doc-only mode. (error: %s)", err)
			return result, nil
		}
		return result, nil
	}

	return ds.fallback.StartDesign(state, request)
}

func (ds *DesignService) Status() (*DesignStatus, error) {
	if ds.openpencilOn {
		return ds.primary.Status()
	}
	return ds.fallback.Status()
}

func (ds *DesignService) AdapterName() string {
	if ds.openpencilOn {
		return DesignModeOpenPencil
	}
	return DesignModeDocOnly
}

func (ds *DesignService) IsOpenPencilEnabled() bool {
	return ds.openpencilOn
}

func SaveDesignState(mode string, fallbackUsed bool) error {
	content := fmt.Sprintf(`{
  "mode": "%s",
  "fallback_used": %t,
  "updated_at": "%s"
}
`, mode, fallbackUsed, NowISO())
	return WriteFile(DesignStateFile, content)
}

func LoadDesignState() (mode string, fallbackUsed bool, err error) {
	data, err := os.ReadFile(DesignStateFile)
	if err != nil {
		return "", false, err
	}
	var ds struct {
		Mode         string `json:"mode"`
		FallbackUsed bool   `json:"fallback_used"`
	}
	if err := json.Unmarshal(data, &ds); err != nil {
		return "", false, err
	}
	return ds.Mode, ds.FallbackUsed, nil
}
