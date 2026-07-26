package domain

const (
	ModeStitch     = "stitch"
	ModeOpenDesign = "opendesign"
	ModeOpenPencil = "openpencil"
	ModeDocOnly    = "doc-only"
)

const (
	Dir                = ".harness/artifacts/design"
	StitchDir          = ".harness/artifacts/design/stitch"
	StitchScreensDir   = ".harness/artifacts/design/stitch/screens"
	StitchExportsDir   = ".harness/artifacts/design/stitch/exports"
	StitchHTMLDir      = ".harness/artifacts/design/stitch/html"
	StitchTaskFile     = ".harness/artifacts/design/stitch/design-task.md"
	StitchDesignMDFile = ".harness/artifacts/design/stitch/DESIGN.md"
	OpenDesignDir      = ".harness/artifacts/design/opendesign"
	OpenDesignTaskFile = ".harness/artifacts/design/opendesign/design-task.md"
	OpenPencilDir      = ".harness/artifacts/design/openpencil"
	ExportsDir         = ".harness/artifacts/design/openpencil/exports"
	TaskFile           = ".harness/artifacts/design/openpencil/design-task.md"
	StateFile          = ".harness/design-state.json"
)

type Result struct {
	Adapter      string
	Mode         string
	FilesCreated []string
	PenFile      string
	TaskFile     string
	Message      string
	FallbackUsed bool
}

type Status struct {
	Adapter              string
	Mode                 string
	Available            bool
	PenFile              string
	HasBrief             bool
	HasFlows             bool
	HasDecisions         bool
	HasPrototype         bool
	HasWireframes        bool
	HasTaskFile          bool
	HasResponsiveQA      bool
	HasRouteInventory    bool
	HasAssetManifest     bool
	HasSourceScreenshots bool
	HasFidelityReport    bool
	GateChecks           []GateCheck
}
