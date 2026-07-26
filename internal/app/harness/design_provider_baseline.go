package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	DesignBaselineDir          = ".harness/artifacts/design/baseline"
	DesignBaselineTaskFile     = ".harness/artifacts/design/baseline/design-task.md"
	DesignRouteInventoryFile   = ".harness/artifacts/design/route-inventory.md"
	DesignAssetManifestFile    = ".harness/artifacts/design/asset-manifest.json"
	DesignSourceScreenshotsDir = ".harness/artifacts/design/source-screenshots"
	DesignFidelityReportFile   = ".harness/artifacts/design/fidelity-report.md"
)

type BaselineWebDesignProvider struct{}

func NewBaselineWebDesignProvider() *BaselineWebDesignProvider { return &BaselineWebDesignProvider{} }

func (p *BaselineWebDesignProvider) Name() string { return "baseline-web" }

func (p *BaselineWebDesignProvider) Detect() DesignProviderDetection {
	profile, err := LoadProjectProfile()
	if err != nil || profile == nil {
		return DesignProviderDetection{Name: p.Name(), Available: false, Reason: "project profile not available"}
	}
	return DesignProviderDetection{Name: p.Name(), Available: profile.ExistingProject && profile.Structure.HasFrontend, Reason: "baseline required for existing frontend projects"}
}

func (p *BaselineWebDesignProvider) Prepare(state *State, request string) (*DesignProviderResult, error) {
	for _, dir := range []string{DesignBaselineDir, DesignSourceScreenshotsDir} {
		if err := ensureDir(dir); err != nil {
			return nil, err
		}
	}
	content := generateBaselineWebTask(state, request)
	if err := WriteFile(DesignBaselineTaskFile, content); err != nil {
		return nil, err
	}
	files := []string{DesignBaselineTaskFile}
	return &DesignProviderResult{Provider: p.Name(), Status: DesignProviderStatusReady, Files: files, Message: "Evidence-first baseline task created before design generation."}, nil
}

func (p *BaselineWebDesignProvider) Generate(state *State, request string) (*DesignProviderResult, error) {
	return p.Prepare(state, request)
}

func (p *BaselineWebDesignProvider) Publish(state *State) (*DesignProviderResult, error) {
	return &DesignProviderResult{Provider: p.Name(), Status: DesignProviderStatusReady, Message: "baseline-web does not publish; it captures source evidence"}, nil
}

func (p *BaselineWebDesignProvider) Verify(state *State) (*DesignProviderResult, error) {
	checks := EvaluateDesignEvidenceGates(state)
	status := DesignProviderStatusReady
	for _, check := range checks {
		if check.Blocking && !check.Pass {
			status = DesignProviderStatusBlocked
			break
		}
	}
	return &DesignProviderResult{Provider: p.Name(), Status: status, Message: RenderDesignGateSummary(checks)}, nil
}

func (p *BaselineWebDesignProvider) Report(state *State) (*DesignProviderResult, error) {
	result, err := p.Verify(state)
	if err != nil {
		return nil, err
	}
	if err := writeDesignProviderReport(filepath.Join(DesignBaselineDir, "baseline-report.md"), p.Name(), result.Status, result.Message, result.Files); err != nil {
		return nil, err
	}
	result.Files = append(result.Files, filepath.Join(DesignBaselineDir, "baseline-report.md"))
	return result, nil
}

type DesignAssetManifest struct {
	Version     string        `json:"version"`
	GeneratedAt string        `json:"generated_at,omitempty"`
	Assets      []DesignAsset `json:"assets"`
}

type DesignAsset struct {
	Path string `json:"path"`
	Type string `json:"type,omitempty"`
	Role string `json:"role,omitempty"`
}

func LoadDesignAssetManifest() (*DesignAssetManifest, error) {
	data, err := os.ReadFile(DesignAssetManifestFile)
	if err != nil {
		return nil, err
	}
	var manifest DesignAssetManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func AssetManifestHasLogo(manifest *DesignAssetManifest) bool {
	if manifest == nil {
		return false
	}
	for _, asset := range manifest.Assets {
		text := strings.ToLower(asset.Path + " " + asset.Type + " " + asset.Role)
		if strings.Contains(text, "logo") {
			return true
		}
	}
	return false
}

func DiscoverBaselineRoutes(profile *ProjectProfile) []string {
	if profile == nil {
		return nil
	}
	routes := map[string]bool{}
	for _, file := range append(profile.FilesScanned, profile.ExistingArtifacts...) {
		path := filepath.ToSlash(file)
		if strings.HasPrefix(path, "src/pages/") {
			route := strings.TrimPrefix(path, "src/pages")
			route = strings.TrimSuffix(route, filepath.Ext(route))
			route = strings.TrimSuffix(route, "/index")
			if route == "" {
				route = "/"
			}
			routes[route] = true
		}
	}
	out := make([]string, 0, len(routes))
	for route := range routes {
		out = append(out, route)
	}
	sort.Strings(out)
	return out
}

func generateBaselineWebTask(state *State, request string) string {
	profile, _ := LoadProjectProfile()
	routes := DiscoverBaselineRoutes(profile)
	var routesSection strings.Builder
	if len(routes) > 0 {
		routesSection.WriteString("\n## Routes detected from project profile\n\n")
		for _, route := range routes {
			routesSection.WriteString(fmt.Sprintf("- `%s`\n", route))
		}
		routesSection.WriteString("\n")
	}

	return mustRenderTemplate("templates/project/harness/design/baseline-web-task.md", RenderVars{
		"request":        requestOrProjectName(state, request),
		"routes_section": routesSection.String(),
	})
}
