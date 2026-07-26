package harness

import tdd "shipwright/internal/tdd/application"

const TDDPolicyJSON = tdd.TDDPolicyJSON
const TDDPolicyMarkdown = tdd.TDDPolicyMarkdown
const TDDPolicyVersion = tdd.TDDPolicyVersion
const TDDReportFile = tdd.TDDReportFile

const (
	TDDModeStrict    = tdd.TDDModeStrict
	TDDModeSuggested = tdd.TDDModeSuggested
	TDDModeNone      = tdd.TDDModeNone
)

type TDDPolicy = tdd.TDDPolicy
type TDDAssessment = tdd.TDDAssessment

func RefreshTDDPolicy() (*TDDPolicy, error) { return tdd.RefreshTDDPolicy() }
func RefreshTDDPolicyFromProfile(profile *ProjectProfile) (*TDDPolicy, error) {
	return tdd.RefreshTDDPolicyFromProfile(profile)
}
func BuildTDDPolicy(profile *ProjectProfile) *TDDPolicy { return tdd.BuildTDDPolicy(profile) }
func SaveTDDPolicy(policy *TDDPolicy) error             { return tdd.SaveTDDPolicy(policy) }
func LoadTDDPolicy() (*TDDPolicy, error)                { return tdd.LoadTDDPolicy() }
func AssessTDDCompliance() TDDAssessment                { return tdd.AssessTDDCompliance() }
func TDDBlockReason() string                            { return tdd.TDDBlockReason() }
func RenderTDDPolicyMarkdown(policy *TDDPolicy) string  { return tdd.RenderTDDPolicyMarkdown(policy) }
func FormatTDDAssessment(assessment TDDAssessment) string {
	return tdd.FormatTDDAssessment(assessment)
}
