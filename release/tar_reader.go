package release

import (
	"encoding/base64"
	"path"

	"github.com/cloudfoundry-incubator/candiedyaml"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmtar "github.com/cloudfoundry/bosh-micro-cli/tar"
)

type tarReader struct {
	tarFilePath string
	fs          boshsys.FileSystem
	extractor   bmtar.Extractor
}

func NewTarReader(tarFilePath string, fs boshsys.FileSystem, extractor bmtar.Extractor) tarReader {
	return tarReader{
		tarFilePath: tarFilePath,
		fs:          fs,
		extractor:   extractor,
	}
}

func (r tarReader) Read() (Release, error) {
	extractedReleasePath, err := r.extractor.Extract(r.tarFilePath)
	if err != nil {
		return Release{}, bosherr.WrapError(err, "Extracting release")
	}

	var manifest Manifest
	releaseManifestPath := path.Join(extractedReleasePath, "release.MF")
	releaseManifestBytes, err := r.fs.ReadFile(releaseManifestPath)
	if err != nil {
		return Release{}, bosherr.WrapError(err, "Reading release manifest")
	}

	err = candiedyaml.Unmarshal(releaseManifestBytes, &manifest)
	if err != nil {
		return Release{}, bosherr.WrapError(err, "Parsing release manifest")
	}

	release := newReleaseFromManifest(manifest)
	return release, nil
}

func (r tarReader) Close() error {
	return nil
}

func newReleaseFromManifest(manifest Manifest) Release {
	return Release{
		Name:    manifest.Name,
		Version: manifest.Version,

		CommitHash:         manifest.CommitHash,
		UncommittedChanges: manifest.UncommittedChanges,

		Jobs:     newJobsFromManifestJobs(manifest.Jobs),
		Packages: newPackagesFromManifestPackages(manifest.Packages),
	}
}

func newJobsFromManifestJobs(manifestJobs []manifestJob) []Job {
	jobs := []Job{}
	for _, manifestJob := range manifestJobs {
		version, _ := decodePossibleBase64Str(manifestJob.Version)
		fingerprint, _ := decodePossibleBase64Str(manifestJob.Fingerprint)
		sha, _ := decodePossibleBase64Str(manifestJob.Sha1)
		job := Job{
			Name:        manifestJob.Name,
			Version:     version,
			Fingerprint: fingerprint,
			Sha1:        sha,
		}
		jobs = append(jobs, job)
	}
	return jobs
}

func newPackagesFromManifestPackages(manifestPackages []manifestPackage) []Package {
	packages := []Package{}
	for _, manifestPackage := range manifestPackages {
		version, _ := decodePossibleBase64Str(manifestPackage.Version)
		fingerprint, _ := decodePossibleBase64Str(manifestPackage.Fingerprint)
		sha, _ := decodePossibleBase64Str(manifestPackage.Sha1)
		pkg := Package{
			Name:         manifestPackage.Name,
			Version:      version,
			Fingerprint:  fingerprint,
			Sha1:         sha,
			Dependencies: manifestPackage.Dependencies,
		}
		packages = append(packages, pkg)
	}
	return packages
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
