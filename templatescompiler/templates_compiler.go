package templatescompiler

import (
	boshblob "github.com/cloudfoundry/bosh-agent/blobstore"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshcmd "github.com/cloudfoundry/bosh-agent/platform/commands"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type TemplatesCompiler interface {
	Compile(jobs []bmrel.Job, deploymentName string, deploymentProperties map[string]interface{}) error
}

type templatesCompiler struct {
	jobRenderer   JobRenderer
	compressor    boshcmd.Compressor
	blobstore     boshblob.Blobstore
	templatesRepo TemplatesRepo
	fs            boshsys.FileSystem
	logger        boshlog.Logger
}

func NewTemplatesCompiler(
	jobRenderer JobRenderer,
	compressor boshcmd.Compressor,
	blobstore boshblob.Blobstore,
	templatesRepo TemplatesRepo,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) TemplatesCompiler {
	return templatesCompiler{
		jobRenderer:   jobRenderer,
		compressor:    compressor,
		blobstore:     blobstore,
		templatesRepo: templatesRepo,
		fs:            fs,
		logger:        logger,
	}
}

func (tc templatesCompiler) Compile(jobs []bmrel.Job, deploymentName string, deploymentProperties map[string]interface{}) error {
	for _, job := range jobs {
		err := tc.compileJob(job, deploymentName, deploymentProperties)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tc templatesCompiler) compileJob(job bmrel.Job, deploymentName string, deploymentProperties map[string]interface{}) error {
	renderedJob, err := tc.jobRenderer.Render(job, deploymentProperties, deploymentName)
	if err != nil {
		return bosherr.WrapErrorf(err, "Rendering templates for job '%s'", job.Name)
	}
	defer renderedJob.DeleteSilently()

	tarballPath, err := tc.compressor.CompressFilesInDir(renderedJob.Path())
	if err != nil {
		return bosherr.WrapError(err, "Compressing rendered job templates")
	}
	defer tc.compressor.CleanUp(tarballPath)

	blobID, blobSHA1, err := tc.blobstore.Create(tarballPath)
	if err != nil {
		return bosherr.WrapError(err, "Creating blob")
	}

	record := TemplateRecord{
		BlobID:   blobID,
		BlobSHA1: blobSHA1,
	}
	err = tc.templatesRepo.Save(job, record)
	if err != nil {
		return bosherr.WrapError(err, "Saving job to templates repo")
	}

	return nil
}
