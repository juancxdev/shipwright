package harness

type GenericExecutorAdapter struct{}

func (GenericExecutorAdapter) Name() string { return ExecutorGeneric }

func (GenericExecutorAdapter) Description() string {
	return "Generic AI-agent bootstrap: AGENTS.md with Shipwright lifecycle rules."
}

func (GenericExecutorAdapter) Generate() (*ExecutorGenerateResult, error) {
	result := &ExecutorGenerateResult{Name: ExecutorGeneric}
	if err := ensureTrackedFile(CommunicationPolicyFile, DefaultCommunicationPolicyMarkdown(), result); err != nil {
		return nil, err
	}
	if err := writeTrackedFile("AGENTS.md", genericAgentsMD(), result); err != nil {
		return nil, err
	}
	result.Message = "Generic executor instructions generated in AGENTS.md."
	return result, nil
}

func (GenericExecutorAdapter) Status() (*ExecutorStatus, error) {
	return requiredStatus(ExecutorGeneric, []string{CommunicationPolicyFile, "AGENTS.md"}), nil
}

const genericAgentsTemplate = "templates/project/harness/executor/generic-agents.md"

func genericAgentsMD() string {
	content, err := RenderTemplate(genericAgentsTemplate, nil)
	if err != nil {
		panic(err)
	}
	return content
}
