package domain

type Pack struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Kind        string    `json:"kind"`
	Description string    `json:"description,omitempty"`
	When        Condition `json:"when"`
	Skills      []string  `json:"skills"`
	Agents      []string  `json:"agents,omitempty"`
}

type Condition struct {
	Stacks            []string `json:"stacks,omitempty"`
	StackNameContains []string `json:"stack_name_contains,omitempty"`
	StackKindContains []string `json:"stack_kind_contains,omitempty"`
	Structure         []string `json:"structure,omitempty"`
	ArtifactsPrefix   []string `json:"artifacts_prefix,omitempty"`
}

type Manifest struct {
	Version     string   `json:"version"`
	GeneratedAt string   `json:"generated_at"`
	Recommended []Pack   `json:"recommended"`
	Installed   []string `json:"installed,omitempty"`
}

type Lock struct {
	Version     string      `json:"version"`
	GeneratedAt string      `json:"generated_at"`
	Packs       []LockPack  `json:"packs"`
	Skills      []LockSkill `json:"skills"`
}

type LockPack struct {
	ID      string   `json:"id"`
	Source  string   `json:"source"`
	Version string   `json:"version"`
	Skills  []string `json:"skills"`
}

type LockSkill struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Source   string `json:"source"`
	Checksum string `json:"checksum"`
}
