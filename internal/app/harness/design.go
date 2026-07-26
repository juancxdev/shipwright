package harness

import (
	"encoding/json"
	"fmt"
	"os"

	designdomain "shipwright/internal/design/domain"
)

const (
	DesignModeStitch     = designdomain.ModeStitch
	DesignModeOpenDesign = designdomain.ModeOpenDesign
	DesignModeOpenPencil = designdomain.ModeOpenPencil
	DesignModeDocOnly    = designdomain.ModeDocOnly
)

const (
	DesignDir                = designdomain.Dir
	DesignStitchDir          = designdomain.StitchDir
	DesignStitchScreensDir   = designdomain.StitchScreensDir
	DesignStitchExportsDir   = designdomain.StitchExportsDir
	DesignStitchHTMLDir      = designdomain.StitchHTMLDir
	DesignStitchTaskFile     = designdomain.StitchTaskFile
	DesignStitchDesignMDFile = designdomain.StitchDesignMDFile
	DesignOpenDesignDir      = designdomain.OpenDesignDir
	DesignOpenDesignTaskFile = designdomain.OpenDesignTaskFile
	DesignOpenPencilDir      = designdomain.OpenPencilDir
	DesignExportsDir         = designdomain.ExportsDir
	DesignTaskFile           = designdomain.TaskFile
	DesignStateFile          = designdomain.StateFile
)

type DesignResult = designdomain.Result

type DesignStatus = designdomain.Status

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
	if err := EnsureDesignEvidenceDirs(); err != nil {
		return nil, err
	}
	baseline := NewBaselineWebDesignProvider()
	baselineResult, err := baseline.Prepare(state, request)
	if err != nil {
		return nil, err
	}

	if ds.primary != ds.fallback {
		result, err := ds.primary.StartDesign(state, request)
		if err != nil {
			result, fbErr := ds.fallback.StartDesign(state, request)
			if fbErr != nil {
				return nil, fmt.Errorf("%s failed: %w; fallback also failed: %v", ds.primary.AdapterName(), err, fbErr)
			}
			result.FallbackUsed = true
			result.FilesCreated = append(baselineResult.Files, result.FilesCreated...)
			result.Message = fmt.Sprintf("Evidence baseline task created. %s unavailable: design generated in doc-only mode. (error: %s)", ds.primary.AdapterName(), err)
			return result, nil
		}
		result.FilesCreated = append(baselineResult.Files, result.FilesCreated...)
		result.Message = "Evidence baseline task created before provider generation. " + result.Message
		return result, nil
	}

	result, err := ds.fallback.StartDesign(state, request)
	if err != nil {
		return nil, err
	}
	result.FilesCreated = append(baselineResult.Files, result.FilesCreated...)
	result.Message = "Evidence baseline task created before doc-only fallback. " + result.Message
	return result, nil
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
