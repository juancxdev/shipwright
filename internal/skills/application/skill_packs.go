package application

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	skillpackdomain "shipwright/internal/skillpacks/domain"
)

const skillPackTemplateRoot = "templates/project/harness/skill-packs"

type SkillPack = skillpackdomain.Pack

type SkillPackCondition = skillpackdomain.Condition

var skillPacks []SkillPack

func init() {
	packs, err := loadSkillPacksFromTemplates()
	if err != nil {
		panic(fmt.Sprintf("cannot load skill pack templates: %v", err))
	}
	skillPacks = packs
}

func loadSkillPacksFromTemplates() ([]SkillPack, error) {
	entries, err := projectTemplateFS.ReadDir(skillPackTemplateRoot)
	if err != nil {
		return nil, err
	}
	packs := make([]SkillPack, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.ToSlash(filepath.Join(skillPackTemplateRoot, entry.Name()))
		data, err := projectTemplateFS.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var pack SkillPack
		if err := json.Unmarshal(data, &pack); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if strings.TrimSpace(pack.ID) == "" || strings.TrimSpace(pack.Name) == "" {
			return nil, fmt.Errorf("%s: skill pack requires id and name", path)
		}
		pack.Skills = sortedUnique(pack.Skills)
		pack.Agents = sortedUnique(pack.Agents)
		packs = append(packs, pack)
	}
	sort.Slice(packs, func(i, j int) bool { return packs[i].ID < packs[j].ID })
	return packs, nil
}

func AllSkillPacks() []SkillPack {
	out := make([]SkillPack, len(skillPacks))
	copy(out, skillPacks)
	return out
}

func MatchingSkillPacks(profile *ProjectProfile) []AssignedTechnology {
	if profile == nil {
		return nil
	}
	var matches []AssignedTechnology
	for _, pack := range skillPacks {
		if evidence := matchSkillPack(pack, profile); len(evidence) > 0 {
			matches = append(matches, AssignedTechnology{
				ID:       pack.ID,
				Name:     pack.Name,
				Kind:     pack.Kind,
				Evidence: sortedUnique(evidence),
				Skills:   sortedUnique(pack.Skills),
			})
		}
	}
	return matches
}

func matchSkillPack(pack SkillPack, profile *ProjectProfile) []string {
	var evidence []string
	if containsStringValue(pack.When.Structure, "frontend") && profile.Structure.HasFrontend {
		evidence = append(evidence, "profile-structure:frontend")
	}
	if containsStringValue(pack.When.Structure, "backend") && profile.Structure.HasBackend {
		evidence = append(evidence, "profile-structure:backend")
	}
	for _, artifact := range profile.ExistingArtifacts {
		normalized := filepath.ToSlash(artifact)
		for _, prefix := range pack.When.ArtifactsPrefix {
			if strings.HasPrefix(normalized, filepath.ToSlash(prefix)) {
				evidence = append(evidence, "profile-artifact:"+normalized)
			}
		}
	}
	for _, stack := range append(profile.Stacks, profile.PlannedStacks...) {
		if skillPackStackMatches(pack.When, stack) {
			evidence = append(evidence, "profile-stack:"+stack.Name)
		}
	}
	return sortedUnique(evidence)
}

func skillPackStackMatches(condition SkillPackCondition, stack StackSignal) bool {
	stackName := strings.ToLower(stack.Name)
	stackKind := strings.ToLower(stack.Kind)
	for _, expected := range condition.Stacks {
		if stackName == strings.ToLower(expected) {
			return true
		}
	}
	for _, needle := range condition.StackNameContains {
		if strings.Contains(stackName, strings.ToLower(needle)) {
			return true
		}
	}
	for _, needle := range condition.StackKindContains {
		if strings.Contains(stackKind, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

const SkillPackManifestJSON = ".harness/skill-packs.json"
const SkillLockJSON = ".harness/skill-lock.json"
const SkillPackInstallRoot = ".harness/skills"
const SkillPackVersion = "1"

type SkillPackManifest = skillpackdomain.Manifest

type SkillLock = skillpackdomain.Lock

type SkillLockPack = skillpackdomain.LockPack

type SkillLockSkill = skillpackdomain.LockSkill

type SkillPackInstallResult struct {
	InstalledPacks  []string `json:"installed_packs"`
	InstalledSkills []string `json:"installed_skills"`
	SkippedSkills   []string `json:"skipped_skills,omitempty"`
	LockPath        string   `json:"lock_path"`
	ManifestPath    string   `json:"manifest_path"`
}

func RecommendedSkillPacks(profile *ProjectProfile) []SkillPack {
	if profile == nil {
		return nil
	}
	var out []SkillPack
	for _, pack := range AllSkillPacks() {
		if evidence := matchSkillPack(pack, profile); len(evidence) > 0 {
			out = append(out, pack)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func SaveRecommendedSkillPackManifest(profile *ProjectProfile) (*SkillPackManifest, error) {
	manifest := &SkillPackManifest{Version: SkillPackVersion, GeneratedAt: nowISO(), Recommended: RecommendedSkillPacks(profile)}
	if lock, err := LoadSkillLock(); err == nil {
		for _, pack := range lock.Packs {
			manifest.Installed = append(manifest.Installed, pack.ID)
		}
		manifest.Installed = sortedUnique(manifest.Installed)
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := writeFile(SkillPackManifestJSON, string(data)+"\n"); err != nil {
		return nil, err
	}
	return manifest, writeFile(".harness/skill-packs.md", RenderSkillPackManifestMarkdown(manifest))
}

func LoadSkillPackManifest() (*SkillPackManifest, error) {
	data, err := os.ReadFile(SkillPackManifestJSON)
	if err != nil {
		return nil, err
	}
	var manifest SkillPackManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func LoadSkillLock() (*SkillLock, error) {
	data, err := os.ReadFile(SkillLockJSON)
	if err != nil {
		return nil, err
	}
	var lock SkillLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, err
	}
	return &lock, nil
}

func InstallRecommendedSkillPacks() (*SkillPackInstallResult, error) {
	profile, err := loadOrCalibrateProjectProfileForAssignments()
	if err != nil {
		return nil, err
	}
	manifest, err := SaveRecommendedSkillPackManifest(profile)
	if err != nil {
		return nil, err
	}
	return InstallSkillPacks(manifest.Recommended)
}

func UpdateInstalledSkillPacks() (*SkillPackInstallResult, error) {
	lock, err := LoadSkillLock()
	if err != nil {
		return InstallRecommendedSkillPacks()
	}
	byID := map[string]SkillPack{}
	for _, pack := range AllSkillPacks() {
		byID[pack.ID] = pack
	}
	var packs []SkillPack
	for _, locked := range lock.Packs {
		if pack, ok := byID[locked.ID]; ok {
			packs = append(packs, pack)
		}
	}
	if len(packs) == 0 {
		return InstallRecommendedSkillPacks()
	}
	return InstallSkillPacks(packs)
}

func InstallSkillPacks(packs []SkillPack) (*SkillPackInstallResult, error) {
	result := &SkillPackInstallResult{LockPath: SkillLockJSON, ManifestPath: SkillPackManifestJSON}
	lock := &SkillLock{Version: SkillPackVersion, GeneratedAt: nowISO()}
	installed := map[string]bool{}
	for _, pack := range packs {
		result.InstalledPacks = append(result.InstalledPacks, pack.ID)
		lock.Packs = append(lock.Packs, SkillLockPack{ID: pack.ID, Source: "shipwright-bundled", Version: SkillPackVersion, Skills: sortedUnique(pack.Skills)})
		for _, skillName := range pack.Skills {
			if installed[skillName] {
				continue
			}
			skill := GetCuratedSkill(skillName)
			if skill == nil {
				result.SkippedSkills = append(result.SkippedSkills, skillName)
				continue
			}
			target := filepath.Join(SkillPackInstallRoot, skill.Name, "SKILL.md")
			if err := writeFile(target, skill.Content); err != nil {
				return result, err
			}
			checksum := checksumString(skill.Content)
			lock.Skills = append(lock.Skills, SkillLockSkill{Name: skill.Name, Path: filepath.ToSlash(target), Source: "shipwright-bundled", Checksum: checksum})
			result.InstalledSkills = append(result.InstalledSkills, skill.Name)
			installed[skillName] = true
		}
	}
	result.InstalledPacks = sortedUnique(result.InstalledPacks)
	result.InstalledSkills = sortedUnique(result.InstalledSkills)
	result.SkippedSkills = sortedUnique(result.SkippedSkills)
	sort.Slice(lock.Packs, func(i, j int) bool { return lock.Packs[i].ID < lock.Packs[j].ID })
	sort.Slice(lock.Skills, func(i, j int) bool { return lock.Skills[i].Name < lock.Skills[j].Name })
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return result, err
	}
	if err := writeFile(SkillLockJSON, string(data)+"\n"); err != nil {
		return result, err
	}
	return result, nil
}

func RenderSkillPackManifestMarkdown(manifest *SkillPackManifest) string {
	var sb strings.Builder
	sb.WriteString("# Skill Packs\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", manifest.GeneratedAt))
	if len(manifest.Recommended) == 0 {
		sb.WriteString("No recommended skill packs detected yet.\n")
		return sb.String()
	}
	sb.WriteString("## Recommended\n\n")
	for _, pack := range manifest.Recommended {
		sb.WriteString(fmt.Sprintf("- `%s` — %s (%d skills)\n", pack.ID, pack.Name, len(pack.Skills)))
	}
	if len(manifest.Installed) > 0 {
		sb.WriteString("\n## Installed\n\n")
		for _, id := range manifest.Installed {
			sb.WriteString(fmt.Sprintf("- `%s`\n", id))
		}
	}
	sb.WriteString("\n## Commands\n\n")
	sb.WriteString("- `shipwright skills install recommended` installs recommended bundled packs into `.harness/skills` and writes `.harness/skill-lock.json`.\n")
	sb.WriteString("- `shipwright skills update` refreshes currently locked bundled packs.\n")
	sb.WriteString("- `shipwright skills import autoskills` remains optional and does not replace the lockfile model.\n")
	return sb.String()
}

func checksumString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

type ExternalSkillInstallResult struct {
	Source          string   `json:"source"`
	InstalledSkills []string `json:"installed_skills"`
	Skipped         []string `json:"skipped,omitempty"`
	LockPath        string   `json:"lock_path"`
}

func InstallSkillsFromSource(source string) (*ExternalSkillInstallResult, error) {
	result := &ExternalSkillInstallResult{Source: source, LockPath: SkillLockJSON}
	source = strings.TrimSpace(source)
	if source == "" {
		return result, fmt.Errorf("source is required")
	}
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		name, content, err := fetchRemoteSkill(source)
		if err != nil {
			return result, err
		}
		if err := installExternalSkill(name, content, source, result); err != nil {
			return result, err
		}
		return result, appendExternalSkillsToLock(result, source)
	}
	info, err := os.Stat(source)
	if err != nil {
		return result, err
	}
	if !info.IsDir() {
		content, err := os.ReadFile(source)
		if err != nil {
			return result, err
		}
		name := inferSkillName(filepath.ToSlash(source))
		if err := installExternalSkill(name, string(content), source, result); err != nil {
			return result, err
		}
		return result, appendExternalSkillsToLock(result, source)
	}
	entries, err := os.ReadDir(source)
	if err != nil {
		return result, err
	}
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") || entry.Name() == "node_modules" {
			continue
		}
		skillFile := filepath.Join(source, entry.Name(), "SKILL.md")
		content, err := os.ReadFile(skillFile)
		if err != nil {
			result.Skipped = append(result.Skipped, entry.Name())
			continue
		}
		if err := installExternalSkill(entry.Name(), string(content), skillFile, result); err != nil {
			return result, err
		}
	}
	result.InstalledSkills = sortedUnique(result.InstalledSkills)
	result.Skipped = sortedUnique(result.Skipped)
	return result, appendExternalSkillsToLock(result, source)
}

func fetchRemoteSkill(source string) (string, string, error) {
	resp, err := http.Get(source)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", "", fmt.Errorf("download %s: %s", source, resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	name := inferSkillName(strings.TrimSuffix(source, "/"))
	frontmatter := parseFrontmatter(string(data))
	if fmName := strings.TrimSpace(frontmatter["name"]); fmName != "" {
		name = fmName
	}
	return name, string(data), nil
}

func installExternalSkill(name, content, source string, result *ExternalSkillInstallResult) error {
	frontmatter := parseFrontmatter(content)
	if fmName := strings.TrimSpace(frontmatter["name"]); fmName != "" {
		name = fmName
	}
	name = normalizeSkillLookup(name)
	if name == "" {
		return fmt.Errorf("cannot infer skill name from %s", source)
	}
	target := filepath.Join(SkillPackInstallRoot, name, "SKILL.md")
	if err := writeFile(target, content); err != nil {
		return err
	}
	result.InstalledSkills = append(result.InstalledSkills, name)
	return nil
}

func appendExternalSkillsToLock(result *ExternalSkillInstallResult, source string) error {
	lock, err := LoadSkillLock()
	if err != nil || lock == nil {
		lock = &SkillLock{Version: SkillPackVersion, GeneratedAt: nowISO()}
	}
	packID := "external-" + normalizeSkillLookup(filepath.Base(strings.TrimSuffix(source, "/")))
	if packID == "external-" {
		packID = "external-skills"
	}
	lock.Packs = append(lock.Packs, SkillLockPack{ID: packID, Source: source, Version: "external", Skills: sortedUnique(result.InstalledSkills)})
	for _, skill := range result.InstalledSkills {
		path := filepath.ToSlash(filepath.Join(SkillPackInstallRoot, skill, "SKILL.md"))
		content, _ := os.ReadFile(path)
		lock.Skills = append(lock.Skills, SkillLockSkill{Name: skill, Path: path, Source: source, Checksum: checksumString(string(content))})
	}
	lock.Packs = uniqueLockPacks(lock.Packs)
	lock.Skills = uniqueLockSkills(lock.Skills)
	lock.GeneratedAt = nowISO()
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(SkillLockJSON, string(data)+"\n")
}

func uniqueLockPacks(packs []SkillLockPack) []SkillLockPack {
	byID := map[string]SkillLockPack{}
	for _, pack := range packs {
		byID[pack.ID] = pack
	}
	out := make([]SkillLockPack, 0, len(byID))
	for _, pack := range byID {
		pack.Skills = sortedUnique(pack.Skills)
		out = append(out, pack)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func uniqueLockSkills(skills []SkillLockSkill) []SkillLockSkill {
	byName := map[string]SkillLockSkill{}
	for _, skill := range skills {
		byName[skill.Name] = skill
	}
	out := make([]SkillLockSkill, 0, len(byName))
	for _, skill := range byName {
		out = append(out, skill)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
