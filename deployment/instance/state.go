package instance

import (
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployment/applyspec"
)

type State interface {
	NetworkInterfaces() []NetworkRef
	RenderedJobs() []JobRef
	CompiledPackages() []PackageRef
	RenderedJobListArchive() BlobRef
	ToApplySpec() bmas.ApplySpec
}

// NetworkRef is a reference to a deployment network, with the interface the instance should use to connect to it.
type NetworkRef struct {
	Name      string
	Interface map[string]interface{}
}

// JobRef is a reference to a rendered job.
// Individual JobRefs do not have Archives because they are aggregated in RenderedJobListArchive.
type JobRef struct {
	Name    string
	Version string
}

// PackageRef is a reference to a compiled package,
type PackageRef struct {
	Name    string
	Version string
	Archive BlobRef
}

// BlobRef is a reference to a file uploaded to the blobstore,
type BlobRef struct {
	BlobstoreID string
	SHA1        string
}

type state struct {
	deploymentName         string
	name                   string
	id                     int
	networks               []NetworkRef
	renderedJobs           []JobRef
	compiledPackages       []PackageRef
	renderedJobListArchive BlobRef
	hash                   string
}

func NewState(
	deploymentName string,
	name string,
	id int,
	networks []NetworkRef,
	renderedJobs []JobRef,
	compiledPackages []PackageRef,
	renderedJobListArchive BlobRef,
	hash string,
) State {
	return &state{
		deploymentName:         deploymentName,
		name:                   name,
		id:                     id,
		networks:               networks,
		renderedJobs:           renderedJobs,
		compiledPackages:       compiledPackages,
		renderedJobListArchive: renderedJobListArchive,
		hash: hash,
	}
}

func (s *state) NetworkInterfaces() []NetworkRef { return s.networks }

func (s *state) RenderedJobs() []JobRef { return s.renderedJobs }

func (s *state) CompiledPackages() []PackageRef { return s.compiledPackages }

func (s *state) RenderedJobListArchive() BlobRef { return s.renderedJobListArchive }

func (s *state) ToApplySpec() bmas.ApplySpec {
	jobTemplateList := make([]bmas.Blob, len(s.renderedJobs), len(s.renderedJobs))
	for i, renderedJob := range s.renderedJobs {
		jobTemplateList[i] = bmas.Blob{
			Name:    renderedJob.Name,
			Version: renderedJob.Version,
		}
	}

	//TODO: once packages are being compiled, create package map based on deployment + releases
	packageMap := make(map[string]bmas.Blob, len(s.compiledPackages))
	for _, compiledPackage := range s.compiledPackages {
		packageMap[compiledPackage.Name] = bmas.Blob{
			Name:        compiledPackage.Name,
			Version:     compiledPackage.Version,
			SHA1:        compiledPackage.Archive.SHA1,
			BlobstoreID: compiledPackage.Archive.BlobstoreID,
		}
	}

	networkMap := make(map[string]interface{}, len(s.networks))
	for _, network := range s.networks {
		networkMap[network.Name] = network.Interface
	}

	return bmas.ApplySpec{
		Deployment: s.deploymentName,
		Index:      s.id,
		Networks:   networkMap,
		Job: bmas.Job{
			Name:      s.name,
			Templates: jobTemplateList,
		},
		Packages: packageMap,
		RenderedTemplatesArchive: bmas.RenderedTemplatesArchiveSpec{
			BlobstoreID: s.renderedJobListArchive.BlobstoreID,
			SHA1:        s.renderedJobListArchive.SHA1,
		},
		ConfigurationHash: s.hash,
	}
}
