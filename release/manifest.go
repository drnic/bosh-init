package release

type Manifest struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`

	CommitHash         string `yaml:"commit_hash"`
	UncommittedChanges bool   `yaml:"uncommitted_changes"`

	Jobs     []manifestJob     `yaml:"jobs"`
	Packages []manifestPackage `yaml:"packages"`
}

type manifestJob struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Fingerprint string `yaml:"fingerprint"`
	Sha1        string `yaml:"sha1"`
}

type manifestPackage struct {
	Name         string   `yaml:"name"`
	Version      string   `yaml:"version"`
	Fingerprint  string   `yaml:"fingerprint"`
	Sha1         string   `yaml:"sha1"`
	Dependencies []string `yaml:"dependencies"`
}
