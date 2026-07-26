package application

import "strings"

type RenderVars map[string]string

func RenderTemplate(path string, vars RenderVars) (string, error) {
	content, err := ReadProjectTemplate(path)
	if err != nil {
		return "", err
	}
	return RenderString(content, vars), nil
}

func RenderString(content string, vars RenderVars) string {
	for key, value := range vars {
		content = strings.ReplaceAll(content, "{{"+key+"}}", value)
	}
	return content
}

func ReadProjectTemplate(path string) (string, error) {
	data, err := projectTemplateFS.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
