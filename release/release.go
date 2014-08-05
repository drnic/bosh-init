package release

type Release struct {
	Name    string
	Version string

	CommitHash         string
	UncommittedChanges bool

	Jobs     []Job
	Packages []Package
}

type Job struct {
	Name        string
	Version     string
	Fingerprint string
	Sha1        string
}

type Package struct {
	Name         string
	Version      string
	Fingerprint  string
	Sha1         string
	Dependencies []string
}
