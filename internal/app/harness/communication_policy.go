package harness

const CommunicationPolicyFile = ".harness/communication-policy.md"

func EnsureCommunicationPolicy() error {
	if ArtifactExists(CommunicationPolicyFile) {
		return nil
	}
	return WriteFile(CommunicationPolicyFile, DefaultCommunicationPolicyMarkdown())
}
