package harness

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	DesignModeStitch     = "stitch"
	DesignModeOpenDesign = "opendesign"
	DesignModeOpenPencil = "openpencil"
	DesignModeDocOnly    = "doc-only"
)

const (
	DesignDir                = ".harness/artifacts/design"
	DesignStitchDir          = ".harness/artifacts/design/stitch"
	DesignStitchScreensDir   = ".harness/artifacts/design/stitch/screens"
	DesignStitchExportsDir   = ".harness/artifacts/design/stitch/exports"
	DesignStitchHTMLDir      = ".harness/artifacts/design/stitch/html"
	DesignStitchTaskFile     = ".harness/artifacts/design/stitch/design-task.md"
	DesignStitchDesignMDFile = ".harness/artifacts/design/stitch/DESIGN.md"
	DesignOpenDesignDir      = ".harness/artifacts/design/opendesign"
	DesignOpenDesignTaskFile = ".harness/artifacts/design/opendesign/design-task.md"
	DesignOpenPencilDir      = ".harness/artifacts/design/openpencil"
	DesignExportsDir         = ".harness/artifacts/design/openpencil/exports"
	DesignTaskFile           = ".harness/artifacts/design/openpencil/design-task.md"
	DesignStateFile          = ".harness/design-state.json"
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
	stitchOn     bool
	opendesignOn bool
	openpencilOn bool
}

func NewDesignService(integrations *Integrations) *DesignService {
	stitchOn := integrations == nil || integrations.Stitch.Enabled
	odOn := integrations != nil && integrations.OpenDesign.Enabled
	opOn := integrations != nil && integrations.OpenPencil.Enabled

	svc := &DesignService{
		fallback:     NewDocOnlyDesignFallback(),
		stitchOn:     stitchOn,
		opendesignOn: odOn,
		openpencilOn: opOn,
	}

	switch {
	case stitchOn:
		svc.primary = NewStitchDesignAdapter()
	case odOn:
		svc.primary = NewOpenDesignAdapter()
	case opOn:
		svc.primary = NewOpenPencilDesignAdapter()
	default:
		svc.primary = svc.fallback
	}

	return svc
}

func (ds *DesignService) StartDesign(state *State, request string) (*DesignResult, error) {
	if ds.primary != ds.fallback {
		result, err := ds.primary.StartDesign(state, request)
		if err != nil {
			result, fbErr := ds.fallback.StartDesign(state, request)
			if fbErr != nil {
				return nil, fmt.Errorf("%s failed: %w; fallback also failed: %v", ds.primary.AdapterName(), err, fbErr)
			}
			result.FallbackUsed = true
			result.Message = fmt.Sprintf("%s unavailable: design generated in doc-only mode. (error: %s)", ds.primary.AdapterName(), err)
			return result, nil
		}
		return result, nil
	}

	return ds.fallback.StartDesign(state, request)
}

func (ds *DesignService) Status() (*DesignStatus, error) {
	return ds.primary.Status()
}

func (ds *DesignService) AdapterName() string {
	return ds.primary.AdapterName()
}

func (ds *DesignService) IsStitchEnabled() bool {
	return ds.stitchOn
}

func (ds *DesignService) IsOpenDesignEnabled() bool {
	return ds.opendesignOn
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
