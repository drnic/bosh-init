package validation

import (
	"fmt"
	"path"

	"github.com/cloudfoundry-incubator/candiedyaml"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type CPIReleaseValidator interface {
	Validate(extractedReleasePath string) error
}

type jobSpec struct {
	Templates map[string]string
}

type cpiReleaseValidator struct {
	fs boshsys.FileSystem
}

func NewCPIReleaseValidator(fs boshsys.FileSystem) cpiReleaseValidator {
	return cpiReleaseValidator{
		fs: fs,
	}
}

func (v cpiReleaseValidator) Validate(extractedReleasePath string) error {
	// validate in spec that bin/cpi exists for some template
	cpiJobPath := path.Join(extractedReleasePath, "jobs", "cpi")
	cpiJobSpec, err := v.loadJobSpec(path.Join(cpiJobPath, "spec"))
	if err != nil {
		return err
	}

	err = v.jobWillRenderPath(cpiJobPath, cpiJobSpec, "bin/cpi")
	if err != nil {
		return err
	}

	return nil
}

func (v cpiReleaseValidator) loadJobSpec(path string) (jobSpec, error) {
	var spec jobSpec

	specBytes, err := v.fs.ReadFile(path)
	if err != nil {
		return spec, bosherr.WrapError(err, "Reading job spec '%s'", path)
	}

	err = candiedyaml.Unmarshal(specBytes, &spec)
	if err != nil {
		return spec, bosherr.WrapError(err, "Unmarshalling job spec '%s'", path)
	}

	return spec, nil
}

func (v cpiReleaseValidator) jobWillRenderPath(jobPath string, spec jobSpec, desiredResultingPath string) error {
	var foundTemplatePath string
	for templatePath, resultingPath := range spec.Templates {
		if resultingPath == desiredResultingPath {
			foundTemplatePath = templatePath
		}
	}
	if foundTemplatePath == "" {
		return fmt.Errorf("No template exists that will render '%s' in job spec %+v (in '%s')", desiredResultingPath, spec, jobPath)
	}
	if !v.fs.FileExists(path.Join(jobPath, "templates", foundTemplatePath)) {
		return fmt.Errorf("'%s' template file '%s' does not exist in job '%s'", desiredResultingPath, foundTemplatePath, jobPath)
	}
	return nil
}
