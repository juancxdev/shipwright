package harness

import templates "shipwright/internal/templates/application"

type TemplateData = templates.TemplateData

func TemplateDataFromState(s *State) TemplateData {
	return TemplateData{
		ProjectName:    s.ProjectName,
		ProjectID:      s.ProjectID,
		InitialRequest: s.InitialRequest,
		CurrentPhase:   s.CurrentPhase,
		RequiresUI:     s.RequiresUI,
	}
}

func GenerateArtifact(path string, data TemplateData) (string, error) {
	return templates.GenerateArtifact(path, data)
}
