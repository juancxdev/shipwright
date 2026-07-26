package domain

type AgentStep struct {
	Title       string
	Description string
}

type Agent struct {
	Name         string
	Filename     string
	Description  string
	Purpose      string
	Inputs       []string
	Outputs      []string
	CanModify    []string
	CanRead      []string
	Steps        []AgentStep
	ReturnFormat string
	DoneCriteria []string
	HandoffRules []string
	Never        []string
}

func (a *Agent) CanModifyArtifact(path string) bool {
	for _, p := range a.CanModify {
		if p == path {
			return true
		}
	}
	return false
}

func (a *Agent) CanReadArtifact(path string) bool {
	for _, p := range a.CanRead {
		if p == path {
			return true
		}
	}
	return a.CanModifyArtifact(path)
}
