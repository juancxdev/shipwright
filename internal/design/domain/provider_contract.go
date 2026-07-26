package domain

const (
	ProviderStatusReady   = "ready"
	ProviderStatusPartial = "partial"
	ProviderStatusBlocked = "blocked"
)

const (
	GateBaselineCaptured  = "baseline-captured"
	GateAssetsPreserved   = "assets-preserved"
	GateProviderPublished = "provider-published"
	GateFidelityVerified  = "fidelity-verified"
	GateTokenQuotaOK      = "token/quota-ok"
)

type ProviderDetection struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
	Mode      string `json:"mode,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

type ProviderResult struct {
	Provider string   `json:"provider"`
	Status   string   `json:"status"`
	Files    []string `json:"files,omitempty"`
	Message  string   `json:"message,omitempty"`
}

type GateCheck struct {
	Gate     string `json:"gate"`
	Pass     bool   `json:"pass"`
	Blocking bool   `json:"blocking"`
	Message  string `json:"message"`
}
