package harness

import skills "shipwright/internal/skills/application"

const AgentCommonFilename = skills.AgentCommonFilename

var AgentCommonProtocol = skills.AgentCommonProtocol

type AgentSkill = skills.AgentSkill

func GetAgentSkill(name string) *AgentSkill   { return skills.GetAgentSkill(name) }
func AllAgentSkills() []AgentSkill            { return skills.AllAgentSkills() }
func GetCuratedSkill(name string) *AgentSkill { return skills.GetCuratedSkill(name) }
func AllCuratedSkills() []AgentSkill          { return skills.AllCuratedSkills() }

const SkillRegistryJSON = skills.SkillRegistryJSON
const SkillRegistryMarkdown = skills.SkillRegistryMarkdown
const SkillRegistryVersion = skills.SkillRegistryVersion

type SkillRegistry = skills.SkillRegistry
type SkillIndex = skills.SkillIndex

func RefreshSkillRegistry() (*SkillRegistry, error)   { return skills.RefreshSkillRegistry() }
func BuildSkillRegistry() (*SkillRegistry, error)     { return skills.BuildSkillRegistry() }
func SaveSkillRegistry(registry *SkillRegistry) error { return skills.SaveSkillRegistry(registry) }
func LoadSkillRegistry() (*SkillRegistry, error)      { return skills.LoadSkillRegistry() }
func RenderSkillRegistryMarkdown(registry *SkillRegistry) string {
	return skills.RenderSkillRegistryMarkdown(registry)
}
func FindSkill(registry *SkillRegistry, name string) *SkillIndex {
	return skills.FindSkill(registry, name)
}

const SkillAssignmentsJSON = skills.SkillAssignmentsJSON
const SkillAssignmentsMarkdown = skills.SkillAssignmentsMarkdown
const SkillAssignmentsVersion = skills.SkillAssignmentsVersion

type SkillAssignmentSet = skills.SkillAssignmentSet
type AssignedTechnology = skills.AssignedTechnology
type AssignedSkill = skills.AssignedSkill
type AgentSkillAssignment = skills.AgentSkillAssignment

func RefreshSkillAssignments() (*SkillAssignmentSet, error) { return skills.RefreshSkillAssignments() }
func RefreshSkillAssignmentsFromRegistry(registry *SkillRegistry) (*SkillAssignmentSet, error) {
	return skills.RefreshSkillAssignmentsFromRegistry(registry)
}
func BuildSkillAssignments(registry *SkillRegistry, profile *ProjectProfile) (*SkillAssignmentSet, error) {
	return skills.BuildSkillAssignments(registry, profile)
}
func SaveSkillAssignments(set *SkillAssignmentSet) error { return skills.SaveSkillAssignments(set) }
func LoadSkillAssignments() (*SkillAssignmentSet, error) { return skills.LoadSkillAssignments() }
func RenderSkillAssignmentsMarkdown(set *SkillAssignmentSet) string {
	return skills.RenderSkillAssignmentsMarkdown(set)
}

const SkillDigestsJSON = skills.SkillDigestsJSON
const SkillDigestsMarkdown = skills.SkillDigestsMarkdown
const SkillDigestsVersion = skills.SkillDigestsVersion

type SkillDigestSet = skills.SkillDigestSet
type AgentSkillDigest = skills.AgentSkillDigest
type DigestSkillRef = skills.DigestSkillRef

func RefreshSkillDigests() (*SkillDigestSet, error) { return skills.RefreshSkillDigests() }
func RefreshSkillDigestsFromRegistry(registry *SkillRegistry) (*SkillDigestSet, error) {
	return skills.RefreshSkillDigestsFromRegistry(registry)
}
func BuildSkillDigests(registry *SkillRegistry, profile *ProjectProfile) *SkillDigestSet {
	return skills.BuildSkillDigests(registry, profile)
}
func SaveSkillDigests(digests *SkillDigestSet) error { return skills.SaveSkillDigests(digests) }
func LoadSkillDigests() (*SkillDigestSet, error)     { return skills.LoadSkillDigests() }
func FindSkillDigest(digests *SkillDigestSet, agent string) *AgentSkillDigest {
	return skills.FindSkillDigest(digests, agent)
}
func RenderSkillDigestsMarkdown(digests *SkillDigestSet) string {
	return skills.RenderSkillDigestsMarkdown(digests)
}

const SkillPackManifestJSON = skills.SkillPackManifestJSON
const SkillLockJSON = skills.SkillLockJSON
const SkillPackInstallRoot = skills.SkillPackInstallRoot
const SkillPackVersion = skills.SkillPackVersion

type SkillPack = skills.SkillPack
type SkillPackCondition = skills.SkillPackCondition
type SkillPackManifest = skills.SkillPackManifest
type SkillLock = skills.SkillLock
type SkillLockPack = skills.SkillLockPack
type SkillLockSkill = skills.SkillLockSkill
type SkillPackInstallResult = skills.SkillPackInstallResult
type ExternalSkillInstallResult = skills.ExternalSkillInstallResult

func AllSkillPacks() []SkillPack { return skills.AllSkillPacks() }
func MatchingSkillPacks(profile *ProjectProfile) []AssignedTechnology {
	return skills.MatchingSkillPacks(profile)
}
func RecommendedSkillPacks(profile *ProjectProfile) []SkillPack {
	return skills.RecommendedSkillPacks(profile)
}
func SaveRecommendedSkillPackManifest(profile *ProjectProfile) (*SkillPackManifest, error) {
	return skills.SaveRecommendedSkillPackManifest(profile)
}
func LoadSkillPackManifest() (*SkillPackManifest, error) { return skills.LoadSkillPackManifest() }
func LoadSkillLock() (*SkillLock, error)                 { return skills.LoadSkillLock() }
func InstallRecommendedSkillPacks() (*SkillPackInstallResult, error) {
	return skills.InstallRecommendedSkillPacks()
}
func UpdateInstalledSkillPacks() (*SkillPackInstallResult, error) {
	return skills.UpdateInstalledSkillPacks()
}
func InstallSkillPacks(packs []SkillPack) (*SkillPackInstallResult, error) {
	return skills.InstallSkillPacks(packs)
}
func RenderSkillPackManifestMarkdown(manifest *SkillPackManifest) string {
	return skills.RenderSkillPackManifestMarkdown(manifest)
}
func InstallSkillsFromSource(source string) (*ExternalSkillInstallResult, error) {
	return skills.InstallSkillsFromSource(source)
}

const AutoSkillsSourceDir = skills.AutoSkillsSourceDir

type AutoSkillsImportResult = skills.AutoSkillsImportResult

func AutoSkillsAvailable() bool { return skills.AutoSkillsAvailable() }
func ImportAutoSkillsToOpenCode() (*AutoSkillsImportResult, error) {
	return skills.ImportAutoSkillsToOpenCode()
}

type PlannedStackRefreshResult = skills.PlannedStackRefreshResult

func RefreshPlannedStackSkillArtifacts() (*PlannedStackRefreshResult, error) {
	return skills.RefreshPlannedStackSkillArtifacts()
}
func RefreshProjectProfileFromPlannedArtifacts() (*ProjectProfile, bool, error) {
	return skills.RefreshProjectProfileFromPlannedArtifacts()
}
func DetectPlannedStacksFromArtifacts() []StackSignal {
	return skills.DetectPlannedStacksFromArtifacts()
}
