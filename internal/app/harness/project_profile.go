package harness

import projectprofile "shipwright/internal/projectprofile/application"

const ProjectProfileJSON = projectprofile.ProjectProfileJSON
const ProjectProfileMarkdown = projectprofile.ProjectProfileMarkdown
const ProjectProfileVersion = projectprofile.ProjectProfileVersion

type ProjectProfile = projectprofile.ProjectProfile
type StackSignal = projectprofile.StackSignal
type ProjectCommands = projectprofile.ProjectCommands
type DetectedCommand = projectprofile.DetectedCommand
type TDDCapability = projectprofile.TDDCapability
type RepositoryProfile = projectprofile.RepositoryProfile
type ProjectStructure = projectprofile.ProjectStructure

func CalibrateProject(projectName string) (*ProjectProfile, error) {
	return projectprofile.CalibrateProject(projectName)
}

func SaveProjectProfile(profile *ProjectProfile) error {
	return projectprofile.SaveProjectProfile(profile)
}

func LoadProjectProfile() (*ProjectProfile, error) {
	return projectprofile.LoadProjectProfile()
}

func RenderProjectProfileMarkdown(profile *ProjectProfile) string {
	return projectprofile.RenderProjectProfileMarkdown(profile)
}
