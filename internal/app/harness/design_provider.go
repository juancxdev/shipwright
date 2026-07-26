package harness

import (
	"fmt"

	designdomain "shipwright/internal/design/domain"
)

type DesignProvider interface {
	Name() string
	Detect() DesignProviderDetection
	Prepare(state *State, request string) (*DesignProviderResult, error)
	Generate(state *State, request string) (*DesignProviderResult, error)
	Publish(state *State) (*DesignProviderResult, error)
	Verify(state *State) (*DesignProviderResult, error)
	Report(state *State) (*DesignProviderResult, error)
}

type DesignProviderDetection = designdomain.ProviderDetection

type DesignProviderResult = designdomain.ProviderResult

const (
	DesignProviderStatusReady   = designdomain.ProviderStatusReady
	DesignProviderStatusPartial = designdomain.ProviderStatusPartial
	DesignProviderStatusBlocked = designdomain.ProviderStatusBlocked
)

func DesignProvidersFor(integrations *Integrations) []DesignProvider {
	providers := []DesignProvider{NewBaselineWebDesignProvider()}
	if integrations != nil && integrations.IsOpenDesignEnabled() {
		providers = append(providers, NewOpenDesignProvider())
	}
	if integrations == nil || integrations.IsStitchEnabled() {
		providers = append(providers, NewStitchProvider())
	}
	if integrations != nil && integrations.IsOpenPencilEnabled() {
		providers = append(providers, NewOpenPencilProvider())
	}
	providers = append(providers, NewDocOnlyProvider())
	return providers
}

func writeDesignProviderReport(path string, provider, status, message string, files []string) error {
	content := fmt.Sprintf(`# Design Provider Report

- Provider: %s
- Status: %s
- Updated: %s

## Message

%s
`, provider, status, NowISO(), message)
	if len(files) > 0 {
		content += "\n## Files\n\n"
		for _, file := range files {
			content += fmt.Sprintf("- `%s`\n", file)
		}
	}
	return WriteFile(path, content)
}
