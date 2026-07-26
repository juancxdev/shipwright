package harness

import (
	"fmt"

	artifactdomain "shipwright/internal/artifacts/domain"
	artifactfs "shipwright/internal/artifacts/fsadapter"
)

const ArtifactRoot = artifactdomain.Root

var artifactDirs = artifactdomain.DefaultLayout().ArtifactDirs
var baseDirs = artifactdomain.DefaultLayout().BaseDirs

func artifactStore() artifactfs.Store {
	return artifactfs.NewStore(artifactdomain.DefaultLayout())
}

func CreateBaseStructure() error {
	if err := artifactStore().CreateBaseStructure(); err != nil {
		return fmt.Errorf("cannot create harness structure: %w", err)
	}
	return nil
}

func ArtifactPath(path string) string {
	return artifactStore().Resolve(path)
}

func ArtifactExists(path string) bool {
	return artifactStore().Exists(path)
}

func CheckArtifacts(paths []string) []string {
	var missing []string
	for _, p := range paths {
		if !ArtifactExists(p) {
			missing = append(missing, p)
		}
	}
	return missing
}

func WriteFile(path, content string) error {
	return artifactStore().Write(path, content)
}

func AppendFile(path, content string) error {
	return artifactStore().Append(path, content)
}
