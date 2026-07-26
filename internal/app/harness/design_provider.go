package harness

import (
	"fmt"
	"strings"

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
	var filesSection strings.Builder
	if len(files) > 0 {
		filesSection.WriteString("\n## Files\n\n")
		for _, file := range files {
			filesSection.WriteString(fmt.Sprintf("- `%s`\n", file))
		}
	}
	content := mustRenderTemplate("templates/project/harness/runtime/design-provider-report.md", RenderVars{
		"provider":      provider,
		"status":        status,
		"updated":       NowISO(),
		"message":       message,
		"files_section": filesSection.String(),
	})
	return WriteFile(path, content)
}
