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

func DefaultCommunicationPolicyMarkdown() string {
	return templates.DefaultCommunicationPolicyMarkdown()
}

type RenderVars = templates.RenderVars

func RenderTemplate(path string, vars RenderVars) (string, error) {
	return templates.RenderTemplate(path, vars)
}

func RenderString(content string, vars RenderVars) string {
	return templates.RenderString(content, vars)
}

func mustRenderTemplate(path string, vars RenderVars) string {
	content, err := RenderTemplate(path, vars)
	if err != nil {
		panic(err)
	}
	return content
}
