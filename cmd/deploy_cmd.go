package cmd

import (
	"errors"
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsys "github.com/cloudfoundry/bosh-agent/system"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmtar "github.com/cloudfoundry/bosh-micro-cli/tar"
	bmui "github.com/cloudfoundry/bosh-micro-cli/ui"
	bmvalidation "github.com/cloudfoundry/bosh-micro-cli/validation"
)

type deployCmd struct {
	ui        bmui.UI
	config    bmconfig.Config
	fs        boshsys.FileSystem
	extractor bmtar.Extractor
}

func NewDeployCmd(
	ui bmui.UI,
	config bmconfig.Config,
	fs boshsys.FileSystem,
	extractor bmtar.Extractor,
) *deployCmd {
	return &deployCmd{
		ui:        ui,
		config:    config,
		fs:        fs,
		extractor: extractor,
	}
}

func (c *deployCmd) Run(args []string) error {
	if len(args) == 0 {
		c.ui.Error("No CPI release provided")
		return errors.New("No CPI release provided")
	}

	if len(c.config.Deployment) == 0 {
		c.ui.Error("No deployment set")
		return errors.New("No deployment set")
	}

	fileValidator := bmvalidation.NewFileValidator(c.fs)
	err := fileValidator.Exists(c.config.Deployment)
	if err != nil {
		c.ui.Error(fmt.Sprintf("Deployment manifest path '%s' does not exist", c.config.Deployment))
		return bosherr.WrapError(err, "Reading deployment manifest for deploy")
	}

	cpiPath := args[0]
	err = fileValidator.Exists(cpiPath)
	if err != nil {
		c.ui.Error(fmt.Sprintf("CPI release '%s' does not exist", cpiPath))
		return bosherr.WrapError(err, "Validating CPI release")
	}

	//releaseReader := release.NewTarReader(cpiPath, c.fs, c.extractor)

	//releaseValidator := bmvalidation.NewCPIReleaseValidator()
	//releaseValidator.Valid(release)

	return nil
}
