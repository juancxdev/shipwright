package application

import (
	"embed"
	"fmt"
	"strings"
)

// projectTemplateFS contains static project templates copied into user projects.
// Keep Markdown/instruction assets here instead of hardcoding long content in Go.
//
//go:embed templates/project/harness
var projectTemplateFS embed.FS

const CommunicationPolicyTemplatePath = "templates/project/harness/communication-policy.md"

func DefaultCommunicationPolicyMarkdown() string {
	data, err := ReadProjectTemplate(CommunicationPolicyTemplatePath)
	if err != nil {
		panic(fmt.Sprintf("cannot load communication policy template: %v", err))
	}
	return data
}

type TemplateData struct {
	ProjectName    string
	ProjectID      string
	InitialRequest string
	CurrentPhase   string
	RequiresUI     *bool
	HasAPI         *bool
}

func GenerateArtifact(path string, data TemplateData) (string, error) {
	templatePath, ok := artifactTemplatePath(path)
	if !ok {
		return "", fmt.Errorf("no template for: %s", path)
	}
	return RenderTemplate(templatePath, RenderVars{
		"project_name":    data.ProjectName,
		"project_id":      data.ProjectID,
		"initial_request": reqStr(data),
		"current_phase":   data.CurrentPhase,
	})
}

func artifactTemplatePath(path string) (string, bool) {
	if !strings.HasPrefix(path, ".harness/artifacts/") {
		return "", false
	}
	templatePath := "templates/project/harness/artifacts/" + strings.TrimPrefix(path, ".harness/artifacts/")
	if _, err := projectTemplateFS.ReadFile(templatePath); err != nil {
		return "", false
	}
	return templatePath, true
}

func reqStr(data TemplateData) string {
	if data.InitialRequest == "" {
		return "(not set)"
	}
	return data.InitialRequest
}
