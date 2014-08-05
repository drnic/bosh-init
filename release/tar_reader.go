package release

import (
	"encoding/base64"
	"path"

	"github.com/cloudfoundry-incubator/candiedyaml"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmerr "github.com/cloudfoundry/bosh-micro-cli/errors"
	bmtar "github.com/cloudfoundry/bosh-micro-cli/tar"
)

type tarReader struct {
	tarFilePath          string
	fs                   boshsys.FileSystem
	extractor            bmtar.Extractor
	extractedReleasePath string
}

func NewTarReader(tarFilePath string, fs boshsys.FileSystem, extractor bmtar.Extractor) *tarReader {
	return &tarReader{
		tarFilePath: tarFilePath,
		fs:          fs,
		extractor:   extractor,
	}
}

func (r *tarReader) Read() (Release, error) {
	var err error
	r.extractedReleasePath, err = r.extractor.Extract(r.tarFilePath)
	if err != nil {
		return Release{}, bosherr.WrapError(err, "Extracting release")
	}

	var manifest Manifest
	releaseManifestPath := path.Join(r.extractedReleasePath, "release.MF")
	releaseManifestBytes, err := r.fs.ReadFile(releaseManifestPath)
	if err != nil {
		return Release{}, bosherr.WrapError(err, "Reading release manifest")
	}

	err = candiedyaml.Unmarshal(releaseManifestBytes, &manifest)
	if err != nil {
		return Release{}, bosherr.WrapError(err, "Parsing release manifest")
	}

	release, err := newReleaseFromManifest(manifest)
	if err != nil {
		return Release{}, bosherr.WrapError(err, "Constructing release from manifest")
	}

	return release, nil
}

func (r *tarReader) Close() error {
	if r.extractedReleasePath != "" {
		return r.fs.RemoveAll(r.extractedReleasePath)
	}

	return nil
}

func newReleaseFromManifest(manifest Manifest) (Release, error) {
	explainableError := bmerr.NewExplainableError("")
	jobs, err := newJobsFromManifestJobs(manifest.Jobs)
	if err != nil {
		explainableError.AddError(bosherr.WrapError(err, "Constructing jobs from manifest"))
	}

	packages, err := newPackagesFromManifestPackages(manifest.Packages)
	if err != nil {
		explainableError.AddError(bosherr.WrapError(err, "Constructing packages from manifest"))
	}

	if explainableError.HasErrors() {
		return Release{}, explainableError
	}

	return Release{
		Name:    manifest.Name,
		Version: manifest.Version,

		CommitHash:         manifest.CommitHash,
		UncommittedChanges: manifest.UncommittedChanges,

		Jobs:     jobs,
		Packages: packages,
	}, nil
}

func newJobsFromManifestJobs(manifestJobs []manifestJob) ([]Job, error) {
	jobs := []Job{}
	explainableError := bmerr.NewExplainableError("")
	for _, manifestJob := range manifestJobs {
		failed := false
		version, err := decodePossibleBase64Str(manifestJob.Version)
		if err != nil {
			explainableError.AddError(bosherr.WrapError(err, "Decoding binary job version for '%s'", manifestJob.Name))
			failed = true
		}
		fingerprint, err := decodePossibleBase64Str(manifestJob.Fingerprint)
		if err != nil {
			explainableError.AddError(bosherr.WrapError(err, "Decoding binary job fingerprint for '%s'", manifestJob.Name))
			failed = true
		}
		sha, err := decodePossibleBase64Str(manifestJob.Sha1)
		if err != nil {
			explainableError.AddError(bosherr.WrapError(err, "Decoding binary job sha for '%s'", manifestJob.Name))
			failed = true
		}

		if !failed {
			job := Job{
				Name:        manifestJob.Name,
				Version:     version,
				Fingerprint: fingerprint,
				Sha1:        sha,
			}
			jobs = append(jobs, job)
		}
	}

	if explainableError.HasErrors() {
		return []Job{}, explainableError
	}

	return jobs, nil
}

func newPackagesFromManifestPackages(manifestPackages []manifestPackage) ([]Package, error) {
	packages := []Package{}
	explainableError := bmerr.NewExplainableError("")
	for _, manifestPackage := range manifestPackages {
		failed := false
		version, err := decodePossibleBase64Str(manifestPackage.Version)
		if err != nil {
			explainableError.AddError(bosherr.WrapError(err, "Decoding binary package version for '%s'", manifestPackage.Name))
			failed = true
		}
		fingerprint, err := decodePossibleBase64Str(manifestPackage.Fingerprint)
		if err != nil {
			explainableError.AddError(bosherr.WrapError(err, "Decoding binary package fingerprint for '%s'", manifestPackage.Name))
			failed = true
		}
		sha, err := decodePossibleBase64Str(manifestPackage.Sha1)
		if err != nil {
			explainableError.AddError(bosherr.WrapError(err, "Decoding binary package sha for '%s'", manifestPackage.Name))
			failed = true
		}

		if !failed {
			pkg := Package{
				Name:        manifestPackage.Name,
				Version:     version,
				Fingerprint: fingerprint,
				Sha1:        sha,

				Dependencies: manifestPackage.Dependencies,
			}
			packages = append(packages, pkg)
		}
	}

	if explainableError.HasErrors() {
		return []Package{}, explainableError
	}

	return packages, nil
}

func decodePossibleBase64Str(str string) (string, error) {
	// Cheating until yaml library provides proper support for !binary
	if str[len(str)-2:] == "==" {
		bytes, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			return "", bosherr.WrapError(err, "Decoding base64 encoded str '%s'", str)
		}

		return string(bytes), nil
	}

	return str, nil
}
