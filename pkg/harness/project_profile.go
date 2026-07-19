package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const ProjectProfileJSON = ".harness/project-profile.json"
const ProjectProfileMarkdown = ".harness/project-profile.md"
const ProjectProfileVersion = "1"

type ProjectProfile struct {
	Version           string            `json:"version"`
	GeneratedAt       string            `json:"generated_at"`
	ProjectName       string            `json:"project_name"`
	Root              string            `json:"root"`
	ExistingProject   bool              `json:"existing_project"`
	Languages         []string          `json:"languages"`
	Stacks            []StackSignal     `json:"stacks"`
	PackageManagers   []string          `json:"package_managers,omitempty"`
	Commands          ProjectCommands   `json:"commands"`
	TDD               TDDCapability     `json:"tdd"`
	Repository        RepositoryProfile `json:"repository"`
	Structure         ProjectStructure  `json:"structure"`
	Conventions       []string          `json:"conventions,omitempty"`
	ExistingArtifacts []string          `json:"existing_artifacts,omitempty"`
	Warnings          []string          `json:"warnings,omitempty"`
	FilesScanned      []string          `json:"files_scanned"`
}

type StackSignal struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Evidence string `json:"evidence"`
}

type ProjectCommands struct {
	Test  []DetectedCommand `json:"test,omitempty"`
	Build []DetectedCommand `json:"build,omitempty"`
	Lint  []DetectedCommand `json:"lint,omitempty"`
	Dev   []DetectedCommand `json:"dev,omitempty"`
}

type DetectedCommand struct {
	Command    string `json:"command"`
	Source     string `json:"source"`
	Confidence string `json:"confidence"`
}

type TDDCapability struct {
	Supported       bool   `json:"supported"`
	RecommendedMode string `json:"recommended_mode"`
	Reason          string `json:"reason"`
}

type RepositoryProfile struct {
	Git           bool     `json:"git"`
	GitIgnore     bool     `json:"gitignore"`
	CI            []string `json:"ci,omitempty"`
	Docker        bool     `json:"docker"`
	DockerCompose bool     `json:"docker_compose"`
}

type ProjectStructure struct {
	MonorepoHints []string `json:"monorepo_hints,omitempty"`
	HasFrontend   bool     `json:"has_frontend"`
	HasBackend    bool     `json:"has_backend"`
	HasSrc        bool     `json:"has_src"`
	HasTests      bool     `json:"has_tests"`
	HasDocs       bool     `json:"has_docs"`
	HasContracts  bool     `json:"has_contracts"`
}

type packageJSON struct {
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Workspaces      any               `json:"workspaces"`
}

func CalibrateProject(projectName string) (*ProjectProfile, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	profile := &ProjectProfile{
		Version:     ProjectProfileVersion,
		GeneratedAt: NowISO(),
		ProjectName: projectName,
		Root:        root,
	}

	knownMarkers := []string{
		"package.json", "go.mod", "pyproject.toml", "requirements.txt", "Cargo.toml", "pom.xml", "build.gradle", "build.gradle.kts", "composer.json", "Gemfile", "pubspec.yaml",
		"frontend", "backend", "apps", "packages", "src", "tests", "test", "__tests__",
	}
	for _, marker := range knownMarkers {
		if pathExists(marker) {
			profile.ExistingProject = true
			break
		}
	}

	detectRepository(profile)
	detectStructure(profile)
	detectNode(profile)
	detectGo(profile)
	detectPython(profile)
	detectRust(profile)
	detectJava(profile)
	detectPHP(profile)
	detectDart(profile)
	detectExistingArtifacts(profile)

	profile.Languages = sortedUnique(profile.Languages)
	profile.PackageManagers = sortedUnique(profile.PackageManagers)
	profile.Conventions = sortedUnique(profile.Conventions)
	profile.ExistingArtifacts = sortedUnique(profile.ExistingArtifacts)
	profile.FilesScanned = sortedUnique(profile.FilesScanned)
	profile.Stacks = uniqueStacks(profile.Stacks)
	profile.Repository.CI = sortedUnique(profile.Repository.CI)
	profile.Structure.MonorepoHints = sortedUnique(profile.Structure.MonorepoHints)
	profile.TDD = inferTDDCapability(profile)
	profile.Warnings = profileWarnings(profile)
	return profile, nil
}

func SaveProjectProfile(profile *ProjectProfile) error {
	if profile == nil {
		return fmt.Errorf("project profile is nil")
	}
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}
	if err := WriteFile(ProjectProfileJSON, string(data)+"\n"); err != nil {
		return err
	}
	return WriteFile(ProjectProfileMarkdown, RenderProjectProfileMarkdown(profile))
}

func LoadProjectProfile() (*ProjectProfile, error) {
	data, err := os.ReadFile(ProjectProfileJSON)
	if err != nil {
		return nil, err
	}
	var profile ProjectProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func RenderProjectProfileMarkdown(profile *ProjectProfile) string {
	var sb strings.Builder
	sb.WriteString("# Project Calibration Profile\n\n")
	sb.WriteString(fmt.Sprintf("**Project:** %s\n", profile.ProjectName))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n", profile.GeneratedAt))
	sb.WriteString(fmt.Sprintf("**Existing project:** %s\n\n", yesNo(profile.ExistingProject)))

	sb.WriteString("## Detected stack\n\n")
	if len(profile.Stacks) == 0 && len(profile.Languages) == 0 {
		sb.WriteString("- No application stack detected yet. Treat this as a greenfield project.\n")
	} else {
		for _, lang := range profile.Languages {
			sb.WriteString(fmt.Sprintf("- Language: `%s`\n", lang))
		}
		for _, stack := range profile.Stacks {
			sb.WriteString(fmt.Sprintf("- %s `%s` — %s\n", stack.Kind, stack.Name, stack.Evidence))
		}
	}

	sb.WriteString("\n## Commands\n\n")
	writeCommandList(&sb, "Test", profile.Commands.Test)
	writeCommandList(&sb, "Build", profile.Commands.Build)
	writeCommandList(&sb, "Lint", profile.Commands.Lint)
	writeCommandList(&sb, "Dev", profile.Commands.Dev)

	sb.WriteString("\n## TDD capability\n\n")
	sb.WriteString(fmt.Sprintf("- Supported: `%s`\n", yesNo(profile.TDD.Supported)))
	sb.WriteString(fmt.Sprintf("- Recommended mode: `%s`\n", profile.TDD.RecommendedMode))
	sb.WriteString(fmt.Sprintf("- Reason: %s\n", profile.TDD.Reason))

	sb.WriteString("\n## Repository & structure\n\n")
	sb.WriteString(fmt.Sprintf("- Git: `%s`\n", yesNo(profile.Repository.Git)))
	sb.WriteString(fmt.Sprintf("- CI: `%s`\n", strings.Join(defaultSlice(profile.Repository.CI, "none"), ", ")))
	sb.WriteString(fmt.Sprintf("- Docker: `%s`\n", yesNo(profile.Repository.Docker)))
	sb.WriteString(fmt.Sprintf("- Docker Compose: `%s`\n", yesNo(profile.Repository.DockerCompose)))
	sb.WriteString(fmt.Sprintf("- Frontend hint: `%s`\n", yesNo(profile.Structure.HasFrontend)))
	sb.WriteString(fmt.Sprintf("- Backend hint: `%s`\n", yesNo(profile.Structure.HasBackend)))
	sb.WriteString(fmt.Sprintf("- Tests directory: `%s`\n", yesNo(profile.Structure.HasTests)))
	if len(profile.Structure.MonorepoHints) > 0 {
		sb.WriteString(fmt.Sprintf("- Monorepo hints: `%s`\n", strings.Join(profile.Structure.MonorepoHints, ", ")))
	}

	if len(profile.Conventions) > 0 {
		sb.WriteString("\n## Detected conventions\n\n")
		for _, convention := range profile.Conventions {
			sb.WriteString(fmt.Sprintf("- %s\n", convention))
		}
	}

	if len(profile.ExistingArtifacts) > 0 {
		sb.WriteString("\n## Existing delivery artifacts\n\n")
		for _, artifact := range profile.ExistingArtifacts {
			sb.WriteString(fmt.Sprintf("- `%s`\n", artifact))
		}
	}

	if len(profile.Warnings) > 0 {
		sb.WriteString("\n## Calibration warnings\n\n")
		for _, warning := range profile.Warnings {
			sb.WriteString(fmt.Sprintf("- %s\n", warning))
		}
	}

	sb.WriteString("\n## How agents must use this\n\n")
	sb.WriteString("- Read this profile before proposing architecture or implementation.\n")
	sb.WriteString("- Prefer detected commands over invented commands.\n")
	sb.WriteString("- If a command is missing, ask or add an explicit task to define it.\n")
	sb.WriteString("- Do not enable strict TDD unless this profile says test capability is present.\n")
	return sb.String()
}

func writeCommandList(sb *strings.Builder, label string, commands []DetectedCommand) {
	if len(commands) == 0 {
		sb.WriteString(fmt.Sprintf("- %s: _not detected_\n", label))
		return
	}
	for _, command := range commands {
		sb.WriteString(fmt.Sprintf("- %s: `%s` (%s, %s)\n", label, command.Command, command.Source, command.Confidence))
	}
}

func detectRepository(profile *ProjectProfile) {
	profile.Repository.Git = dirExists(".git")
	profile.Repository.GitIgnore = pathExists(".gitignore")
	profile.Repository.Docker = pathExists("Dockerfile") || pathExists("dockerfile")
	profile.Repository.DockerCompose = pathExists("docker-compose.yml") || pathExists("docker-compose.yaml") || pathExists("compose.yml") || pathExists("compose.yaml")
	if dirExists(filepath.Join(".github", "workflows")) {
		profile.Repository.CI = append(profile.Repository.CI, "github-actions")
		profile.FilesScanned = append(profile.FilesScanned, filepath.Join(".github", "workflows"))
	}
	if pathExists(".gitlab-ci.yml") {
		profile.Repository.CI = append(profile.Repository.CI, "gitlab-ci")
		profile.FilesScanned = append(profile.FilesScanned, ".gitlab-ci.yml")
	}
}

func detectStructure(profile *ProjectProfile) {
	profile.Structure.HasFrontend = dirExists("frontend") || dirExists(filepath.Join("apps", "web")) || dirExists(filepath.Join("apps", "frontend"))
	profile.Structure.HasBackend = dirExists("backend") || dirExists(filepath.Join("apps", "api")) || dirExists(filepath.Join("apps", "backend"))
	profile.Structure.HasSrc = dirExists("src")
	profile.Structure.HasTests = dirExists("tests") || dirExists("test") || dirExists("__tests__")
	profile.Structure.HasDocs = dirExists("docs")
	profile.Structure.HasContracts = pathExists(filepath.Join("contracts", "openapi.yaml")) || pathExists(filepath.Join("contracts", "openapi.yml")) || pathExists("openapi.yaml") || pathExists("openapi.yml")
	if dirExists("apps") {
		profile.Structure.MonorepoHints = append(profile.Structure.MonorepoHints, "apps/")
	}
	if dirExists("packages") {
		profile.Structure.MonorepoHints = append(profile.Structure.MonorepoHints, "packages/")
	}
	if profile.Structure.HasFrontend {
		profile.Conventions = append(profile.Conventions, "frontend code appears separated from backend")
	}
	if profile.Structure.HasBackend {
		profile.Conventions = append(profile.Conventions, "backend code appears separated from frontend")
	}
	if profile.Structure.HasTests {
		profile.Conventions = append(profile.Conventions, "test directory exists")
	}
}

func detectNode(profile *ProjectProfile) {
	if !pathExists("package.json") {
		return
	}
	profile.FilesScanned = append(profile.FilesScanned, "package.json")
	profile.Languages = append(profile.Languages, "JavaScript")
	profile.PackageManagers = append(profile.PackageManagers, detectNodePackageManager())
	profile.Stacks = append(profile.Stacks, StackSignal{Name: "Node.js", Kind: "runtime", Evidence: "package.json"})
	if pathExists("tsconfig.json") {
		profile.Languages = append(profile.Languages, "TypeScript")
		profile.FilesScanned = append(profile.FilesScanned, "tsconfig.json")
		profile.Conventions = append(profile.Conventions, "TypeScript config present")
	}
	data, err := os.ReadFile("package.json")
	if err != nil {
		profile.Warnings = append(profile.Warnings, "package.json exists but could not be read")
		return
	}
	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		profile.Warnings = append(profile.Warnings, "package.json exists but is not valid JSON")
		return
	}
	pm := firstOrDefault(profile.PackageManagers, "npm")
	addNpmScriptCommands(profile, pm, pkg.Scripts)
	detectNodeFrameworks(profile, pkg)
	if pkg.Workspaces != nil || pathExists("pnpm-workspace.yaml") || pathExists("turbo.json") || pathExists("nx.json") {
		profile.Structure.MonorepoHints = append(profile.Structure.MonorepoHints, "node-workspaces")
	}
}

func detectNodePackageManager() string {
	switch {
	case pathExists("pnpm-lock.yaml"):
		return "pnpm"
	case pathExists("yarn.lock"):
		return "yarn"
	case pathExists("bun.lock") || pathExists("bun.lockb"):
		return "bun"
	case pathExists("package-lock.json"):
		return "npm"
	default:
		return "npm"
	}
}

func addNpmScriptCommands(profile *ProjectProfile, pm string, scripts map[string]string) {
	if len(scripts) == 0 {
		return
	}
	if script, ok := scripts["test"]; ok && isUsefulNpmScript(script) {
		profile.Commands.Test = append(profile.Commands.Test, DetectedCommand{Command: pm + " test", Source: "package.json scripts.test", Confidence: "high"})
	}
	if _, ok := scripts["build"]; ok {
		profile.Commands.Build = append(profile.Commands.Build, DetectedCommand{Command: pm + " run build", Source: "package.json scripts.build", Confidence: "high"})
	}
	if _, ok := scripts["lint"]; ok {
		profile.Commands.Lint = append(profile.Commands.Lint, DetectedCommand{Command: pm + " run lint", Source: "package.json scripts.lint", Confidence: "high"})
	}
	if _, ok := scripts["dev"]; ok {
		profile.Commands.Dev = append(profile.Commands.Dev, DetectedCommand{Command: pm + " run dev", Source: "package.json scripts.dev", Confidence: "high"})
	}
}

func isUsefulNpmScript(script string) bool {
	lower := strings.ToLower(script)
	return strings.TrimSpace(script) != "" && !strings.Contains(lower, "no test specified") && !strings.Contains(lower, "exit 1")
}

func detectNodeFrameworks(profile *ProjectProfile, pkg packageJSON) {
	deps := map[string]string{}
	for k, v := range pkg.Dependencies {
		deps[k] = v
	}
	for k, v := range pkg.DevDependencies {
		deps[k] = v
	}
	frameworks := map[string]string{
		"next":             "Next.js",
		"react":            "React",
		"vue":              "Vue",
		"@angular/core":    "Angular",
		"svelte":           "Svelte",
		"vite":             "Vite",
		"express":          "Express",
		"@nestjs/core":     "NestJS",
		"fastify":          "Fastify",
		"vitest":           "Vitest",
		"jest":             "Jest",
		"@playwright/test": "Playwright",
	}
	for dep, name := range frameworks {
		if _, ok := deps[dep]; ok {
			kind := "framework"
			if name == "Vitest" || name == "Jest" || name == "Playwright" {
				kind = "test-tool"
			}
			profile.Stacks = append(profile.Stacks, StackSignal{Name: name, Kind: kind, Evidence: "package.json dependency " + dep})
		}
	}
}

func detectGo(profile *ProjectProfile) {
	if !pathExists("go.mod") {
		return
	}
	profile.FilesScanned = append(profile.FilesScanned, "go.mod")
	profile.Languages = append(profile.Languages, "Go")
	profile.Stacks = append(profile.Stacks, StackSignal{Name: "Go modules", Kind: "build-system", Evidence: "go.mod"})
	profile.Commands.Test = append(profile.Commands.Test, DetectedCommand{Command: "go test ./...", Source: "go.mod", Confidence: "high"})
	profile.Commands.Build = append(profile.Commands.Build, DetectedCommand{Command: "go build ./...", Source: "go.mod", Confidence: "medium"})
	if dirExists("cmd") {
		profile.Conventions = append(profile.Conventions, "Go cmd/ entrypoint directory present")
	}
	if dirExists("internal") {
		profile.Conventions = append(profile.Conventions, "Go internal/ package boundary present")
	}
}

func detectPython(profile *ProjectProfile) {
	if !pathExists("pyproject.toml") && !pathExists("requirements.txt") && !pathExists("Pipfile") && !pathExists("poetry.lock") && !pathExists("uv.lock") {
		return
	}
	profile.Languages = append(profile.Languages, "Python")
	profile.Stacks = append(profile.Stacks, StackSignal{Name: "Python", Kind: "runtime", Evidence: pythonEvidence(profile)})
	if pathExists("pyproject.toml") {
		profile.FilesScanned = append(profile.FilesScanned, "pyproject.toml")
	}
	if pathExists("requirements.txt") {
		profile.FilesScanned = append(profile.FilesScanned, "requirements.txt")
	}
	if fileContainsAny("pyproject.toml", []string{"pytest", "unittest"}) || fileContainsAny("requirements.txt", []string{"pytest"}) || profile.Structure.HasTests {
		profile.Commands.Test = append(profile.Commands.Test, DetectedCommand{Command: "pytest", Source: "python project files", Confidence: "medium"})
	}
}

func pythonEvidence(profile *ProjectProfile) string {
	for _, file := range []string{"pyproject.toml", "requirements.txt", "Pipfile", "poetry.lock", "uv.lock"} {
		if pathExists(file) {
			return file
		}
	}
	return "python files"
}

func detectRust(profile *ProjectProfile) {
	if !pathExists("Cargo.toml") {
		return
	}
	profile.FilesScanned = append(profile.FilesScanned, "Cargo.toml")
	profile.Languages = append(profile.Languages, "Rust")
	profile.PackageManagers = append(profile.PackageManagers, "cargo")
	profile.Stacks = append(profile.Stacks, StackSignal{Name: "Cargo", Kind: "build-system", Evidence: "Cargo.toml"})
	profile.Commands.Test = append(profile.Commands.Test, DetectedCommand{Command: "cargo test", Source: "Cargo.toml", Confidence: "high"})
	profile.Commands.Build = append(profile.Commands.Build, DetectedCommand{Command: "cargo build", Source: "Cargo.toml", Confidence: "high"})
}

func detectJava(profile *ProjectProfile) {
	switch {
	case pathExists("pom.xml"):
		profile.FilesScanned = append(profile.FilesScanned, "pom.xml")
		profile.Languages = append(profile.Languages, "Java")
		profile.PackageManagers = append(profile.PackageManagers, "maven")
		profile.Stacks = append(profile.Stacks, StackSignal{Name: "Maven", Kind: "build-system", Evidence: "pom.xml"})
		profile.Commands.Test = append(profile.Commands.Test, DetectedCommand{Command: "mvn test", Source: "pom.xml", Confidence: "high"})
		profile.Commands.Build = append(profile.Commands.Build, DetectedCommand{Command: "mvn package", Source: "pom.xml", Confidence: "medium"})
	case pathExists("build.gradle") || pathExists("build.gradle.kts"):
		file := "build.gradle"
		if pathExists("build.gradle.kts") {
			file = "build.gradle.kts"
		}
		profile.FilesScanned = append(profile.FilesScanned, file)
		profile.Languages = append(profile.Languages, "Java")
		profile.PackageManagers = append(profile.PackageManagers, "gradle")
		profile.Stacks = append(profile.Stacks, StackSignal{Name: "Gradle", Kind: "build-system", Evidence: file})
		profile.Commands.Test = append(profile.Commands.Test, DetectedCommand{Command: "./gradlew test", Source: file, Confidence: "medium"})
		profile.Commands.Build = append(profile.Commands.Build, DetectedCommand{Command: "./gradlew build", Source: file, Confidence: "medium"})
	}
}

func detectPHP(profile *ProjectProfile) {
	if !pathExists("composer.json") {
		return
	}
	profile.FilesScanned = append(profile.FilesScanned, "composer.json")
	profile.Languages = append(profile.Languages, "PHP")
	profile.PackageManagers = append(profile.PackageManagers, "composer")
	profile.Stacks = append(profile.Stacks, StackSignal{Name: "Composer", Kind: "build-system", Evidence: "composer.json"})
	if fileContainsAny("composer.json", []string{"phpunit", "pest"}) {
		profile.Commands.Test = append(profile.Commands.Test, DetectedCommand{Command: "composer test", Source: "composer.json", Confidence: "medium"})
	}
}

func detectDart(profile *ProjectProfile) {
	if !pathExists("pubspec.yaml") {
		return
	}
	profile.FilesScanned = append(profile.FilesScanned, "pubspec.yaml")
	profile.Languages = append(profile.Languages, "Dart")
	profile.PackageManagers = append(profile.PackageManagers, "pub")
	profile.Stacks = append(profile.Stacks, StackSignal{Name: "Dart/Flutter", Kind: "runtime", Evidence: "pubspec.yaml"})
	profile.Commands.Test = append(profile.Commands.Test, DetectedCommand{Command: "dart test", Source: "pubspec.yaml", Confidence: "medium"})
}

func detectExistingArtifacts(profile *ProjectProfile) {
	for _, artifact := range []string{
		"README.md", "docs", "contracts/openapi.yaml", "contracts/openapi.yml", "openapi.yaml", "openapi.yml", ".opencode/opencode.json", "AGENTS.md",
	} {
		if pathExists(artifact) || dirExists(artifact) {
			profile.ExistingArtifacts = append(profile.ExistingArtifacts, artifact)
		}
	}
}

func inferTDDCapability(profile *ProjectProfile) TDDCapability {
	if len(profile.Commands.Test) > 0 {
		return TDDCapability{
			Supported:       true,
			RecommendedMode: "strict",
			Reason:          "test command detected; agents can run red/green verification before implementation",
		}
	}
	if len(profile.Languages) > 0 {
		return TDDCapability{
			Supported:       false,
			RecommendedMode: "suggested",
			Reason:          "stack detected but no reliable test command found; define tests before enforcing strict TDD",
		}
	}
	return TDDCapability{
		Supported:       false,
		RecommendedMode: "none",
		Reason:          "greenfield project with no stack/test runner detected yet",
	}
}

func profileWarnings(profile *ProjectProfile) []string {
	var warnings []string
	if profile.ExistingProject && len(profile.Commands.Test) == 0 {
		warnings = append(warnings, "existing project detected but no test command found")
	}
	if profile.ExistingProject && !profile.Repository.Git {
		warnings = append(warnings, "existing project is not a git repository or .git is not visible")
	}
	if len(profile.Languages) > 1 {
		warnings = append(warnings, "multiple languages detected; confirm frontend/backend boundaries before implementation")
	}
	return warnings
}

func pathExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileContainsAny(path string, needles []string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	lower := strings.ToLower(string(data))
	for _, needle := range needles {
		if strings.Contains(lower, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func sortedUnique(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func uniqueStacks(values []StackSignal) []StackSignal {
	seen := map[string]bool{}
	var out []StackSignal
	for _, value := range values {
		key := value.Kind + ":" + value.Name + ":" + value.Evidence
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind == out[j].Kind {
			return out[i].Name < out[j].Name
		}
		return out[i].Kind < out[j].Kind
	})
	return out
}

func firstOrDefault(values []string, fallback string) string {
	if len(values) == 0 {
		return fallback
	}
	return values[0]
}

func defaultSlice(values []string, fallback string) []string {
	if len(values) == 0 {
		return []string{fallback}
	}
	return values
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
