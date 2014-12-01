package validator

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmcpi "github.com/cloudfoundry/bosh-micro-cli/cpi"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmdeplval "github.com/cloudfoundry/bosh-micro-cli/deployment/validator"
	bmenventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type Validator interface {
	Start()
	ValidateDeployment(
		deploymentFile string,
		deploymentParser bmdepl.Parser,
		boshDeploymentValidator bmdeplval.DeploymentValidator,
	) (bmdepl.Deployment, bmdepl.CPIDeployment, error)
	ValidateRelease(cpiReleaseTarballPath string, cpiInstaller bmcpi.Installer) (bmrel.Release, error)
	ValidateStemcell(stemcellTarballPath string, stemcellExtractor bmstemcell.Extractor) (bmstemcell.ExtractedStemcell, error)
	Finish()
}

type validator struct {
	eventLogger bmenventlog.EventLogger
	fs          boshsys.FileSystem
	logger      boshlog.Logger
	logTag      string

	validationStage bmenventlog.Stage
}

func NewValidator(
	eventLogger bmenventlog.EventLogger,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) *validator {
	return &validator{
		eventLogger: eventLogger,
		fs:          fs,
		logger:      logger,
		logTag:      "validator",
	}
}

func (v *validator) Start() {
	v.validationStage = v.eventLogger.NewStage("validating")
	v.validationStage.Start()
}

func (v *validator) ValidateDeployment(
	deploymentFile string,
	deploymentParser bmdepl.Parser,
	boshDeploymentValidator bmdeplval.DeploymentValidator,
) (bmdepl.Deployment, bmdepl.CPIDeployment, error) {
	manifestValidationStep := v.validationStage.NewStep("Validating deployment manifest")
	manifestValidationStep.Start()

	if deploymentFile == "" {
		err := bosherr.New("No deployment set")
		manifestValidationStep.Fail(err.Error())
		return bmdepl.Deployment{}, bmdepl.CPIDeployment{}, err
	}

	v.logger.Info(v.logTag, "Checking for deployment `%s'", deploymentFile)
	if !v.fs.FileExists(deploymentFile) {
		err := bosherr.New("Verifying that the deployment `%s' exists", deploymentFile)
		manifestValidationStep.Fail(err.Error())
		return bmdepl.Deployment{}, bmdepl.CPIDeployment{}, err
	}

	boshDeployment, cpiDeployment, err := deploymentParser.Parse(deploymentFile)
	if err != nil {
		err = bosherr.WrapError(err, "Parsing deployment manifest `%s'", deploymentFile)
		manifestValidationStep.Fail(err.Error())
		return bmdepl.Deployment{}, bmdepl.CPIDeployment{}, err
	}

	err = boshDeploymentValidator.Validate(boshDeployment)
	if err != nil {
		err = bosherr.WrapError(err, "Validating deployment manifest")
		manifestValidationStep.Fail(err.Error())
		return bmdepl.Deployment{}, bmdepl.CPIDeployment{}, err
	}

	manifestValidationStep.Finish()

	return boshDeployment, cpiDeployment, nil
}

func (v *validator) ValidateRelease(cpiReleaseTarballPath string, cpiInstaller bmcpi.Installer) (bmrel.Release, error) {
	cpiValidationStep := v.validationStage.NewStep("Validating cpi release")
	cpiValidationStep.Start()

	if !v.fs.FileExists(cpiReleaseTarballPath) {
		err := bosherr.New("Verifying that the CPI release `%s' exists", cpiReleaseTarballPath)
		cpiValidationStep.Fail(err.Error())
		return nil, err
	}

	cpiRelease, err := cpiInstaller.Extract(cpiReleaseTarballPath)
	if err != nil {
		err = bosherr.WrapError(err, "Extracting CPI release `%s'", cpiReleaseTarballPath)
		cpiValidationStep.Fail(err.Error())
		return nil, err
	}

	cpiValidationStep.Finish()

	return cpiRelease, nil
}

func (v *validator) ValidateStemcell(stemcellTarballPath string, stemcellExtractor bmstemcell.Extractor) (bmstemcell.ExtractedStemcell, error) {
	stemcellValidationStep := v.validationStage.NewStep("Validating stemcell")
	stemcellValidationStep.Start()

	if !v.fs.FileExists(stemcellTarballPath) {
		err := bosherr.New("Verifying that the stemcell `%s' exists", stemcellTarballPath)
		stemcellValidationStep.Fail(err.Error())
		return nil, err
	}

	extractedStemcell, err := stemcellExtractor.Extract(stemcellTarballPath)
	if err != nil {
		err = bosherr.WrapError(err, "Extracting stemcell from `%s'", stemcellTarballPath)
		stemcellValidationStep.Fail(err.Error())
		return nil, err
	}

	stemcellValidationStep.Finish()

	return extractedStemcell, nil
}

func (v *validator) Finish() {
	v.validationStage.Finish()
	v.validationStage = nil
}
