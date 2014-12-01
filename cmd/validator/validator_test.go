package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmcpi "github.com/cloudfoundry/bosh-micro-cli/cpi/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell/fakes"
	fakebmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment/fakes"
	fakebmdeplval "github.com/cloudfoundry/bosh-micro-cli/deployment/validator/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/cmd/validator"
)

var _ = Describe("Validator", func() {
	var (
		validator Validator
		fakeFs    *fakesys.FakeFileSystem
		logger    boshlog.Logger

		fakeDeploymentParser    *fakebmdepl.FakeParser
		fakeDeploymentValidator *fakebmdeplval.FakeValidator

		fakeEventLogger *fakebmlog.FakeEventLogger
		fakeStage       *fakebmlog.FakeStage
	)

	BeforeEach(func() {
		fakeEventLogger = fakebmlog.NewFakeEventLogger()
		fakeStage = fakebmlog.NewFakeStage()
		fakeEventLogger.SetNewStageBehavior(fakeStage)
		fakeFs = fakesys.NewFakeFileSystem()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		validator = NewValidator(fakeEventLogger, fakeFs, logger)
	})

	Describe("Start", func() {
		It("adds a new event logger stage", func() {
			validator.Start()
			Expect(fakeEventLogger.NewStageInputs).To(Equal([]fakebmlog.NewStageInput{
				{
					Name: "validating",
				},
			}))

			Expect(fakeStage.Started).To(BeTrue())
		})
	})

	Describe("ValidateDeployment", func() {
		var (
			expectedBoshDeployment bmdepl.Deployment
			expectedCPIDeployment  bmdepl.CPIDeployment
		)

		BeforeEach(func() {
			fakeDeploymentParser = fakebmdepl.NewFakeParser()
			fakeFs.WriteFileString("fake-deployment-manifest", "")
			expectedCPIDeployment = bmdepl.CPIDeployment{
				Registry: bmdepl.Registry{
					Username: "fake-username",
				},
				SSHTunnel: bmdepl.SSHTunnel{
					Host: "fake-host",
				},
				Mbus: "http://fake-mbus-user:fake-mbus-password@fake-mbus-endpoint",
			}
			fakeDeploymentParser.ParseCPIDeployment = expectedCPIDeployment

			expectedBoshDeployment = bmdepl.Deployment{
				Name: "fake-deployment-name",
				Jobs: []bmdepl.Job{
					{
						Name: "fake-job-name",
					},
				},
			}
			fakeDeploymentParser.ParseDeployment = expectedBoshDeployment

			fakeDeploymentValidator = fakebmdeplval.NewFakeValidator()
			fakeDeploymentValidator.SetValidateBehavior([]fakebmdeplval.ValidateOutput{
				{
					Err: nil,
				},
			})

			validator.Start()
		})

		It("parses the deployment manifest", func() {
			deployment, cpiDeployment, err := validator.ValidateDeployment("fake-deployment-manifest", fakeDeploymentParser, fakeDeploymentValidator)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDeploymentParser.ParsePath).To(Equal("fake-deployment-manifest"))
			Expect(deployment).To(Equal(expectedBoshDeployment))
			Expect(cpiDeployment).To(Equal(expectedCPIDeployment))
		})

		It("validates bosh deployment manifest", func() {
			_, _, err := validator.ValidateDeployment("fake-deployment-manifest", fakeDeploymentParser, fakeDeploymentValidator)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeDeploymentValidator.ValidateInputs).To(Equal([]fakebmdeplval.ValidateInput{
				{
					Deployment: expectedBoshDeployment,
				},
			}))
		})

		It("logs validation stages", func() {
			_, _, err := validator.ValidateDeployment("fake-deployment-manifest", fakeDeploymentParser, fakeDeploymentValidator)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Validating deployment manifest",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
		})

		Context("when deployment file does not exist", func() {
			BeforeEach(func() {
				fakeFs.RemoveAll("fake-deployment-manifest")
			})

			It("returns error", func() {
				_, _, err := validator.ValidateDeployment("fake-deployment-manifest", fakeDeploymentParser, fakeDeploymentValidator)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Verifying that the deployment `fake-deployment-manifest' exists"))

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Validating deployment manifest",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Verifying that the deployment `fake-deployment-manifest' exists",
				}))
			})
		})
	})

	Describe("ValidateRelease", func() {
		var (
			expectedRelease  bmrel.Release
			fakeCPIInstaller *fakebmcpi.FakeInstaller
		)

		BeforeEach(func() {
			fakeCPIInstaller = fakebmcpi.NewFakeInstaller()
			expectedRelease = bmrel.NewRelease(
				"fake-release",
				"fake-version",
				[]bmrel.Job{},
				[]*bmrel.Package{},
				"/some/release/path",
				fakeFs,
			)
			fakeFs.WriteFileString("fake-release-path", "")
			fakeCPIInstaller.SetExtractBehavior("fake-release-path", expectedRelease, nil)
			validator.Start()
		})

		It("extracts CPI release tarball", func() {
			release, err := validator.ValidateRelease("fake-release-path", fakeCPIInstaller)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCPIInstaller.ExtractInputs).To(Equal([]fakebmcpi.ExtractInput{
				{
					ReleaseTarballPath: "fake-release-path",
				},
			}))

			Expect(release).To(Equal(expectedRelease))
		})

		It("logs validation stages", func() {
			_, err := validator.ValidateRelease("fake-release-path", fakeCPIInstaller)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Validating cpi release",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))
		})

		Context("When the CPI release tarball does not exist", func() {
			BeforeEach(func() {
				fakeFs.RemoveAll("fake-release-path")
			})

			It("returns error", func() {
				_, err := validator.ValidateRelease("fake-release-path", fakeCPIInstaller)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Verifying that the CPI release `fake-release-path' exists"))

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Validating cpi release",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Verifying that the CPI release `fake-release-path' exists",
				}))
			})
		})
	})

	Describe("ValidateStemcell", func() {
		var (
			expectedStemcell      bmstemcell.ExtractedStemcell
			fakeStemcellExtractor *fakebmstemcell.FakeExtractor
		)

		BeforeEach(func() {
			fakeStemcellExtractor = fakebmstemcell.NewFakeExtractor()
			expectedStemcell = bmstemcell.NewExtractedStemcell(
				bmstemcell.Manifest{
					ImagePath:          "/stemcell/image/path",
					Name:               "fake-stemcell-name",
					Version:            "fake-stemcell-version",
					SHA1:               "fake-stemcell-sha1",
					RawCloudProperties: map[interface{}]interface{}{},
				},
				bmstemcell.ApplySpec{},
				"fake-extracted-path",
				fakeFs,
			)
			fakeStemcellExtractor.SetExtractBehavior("fake-stemcell-path", expectedStemcell, nil)
			fakeFs.WriteFileString("fake-stemcell-path", "")
			validator.Start()
		})

		It("logs validation stages", func() {
			stemcell, err := validator.ValidateStemcell("fake-stemcell-path", fakeStemcellExtractor)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
				Name: "Validating stemcell",
				States: []bmeventlog.EventState{
					bmeventlog.Started,
					bmeventlog.Finished,
				},
			}))

			Expect(stemcell).To(Equal(expectedStemcell))
		})

		It("extracts the stemcell", func() {
			_, err := validator.ValidateStemcell("fake-stemcell-path", fakeStemcellExtractor)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeStemcellExtractor.ExtractInputs).To(Equal([]fakebmstemcell.ExtractInput{
				{
					TarballPath: "fake-stemcell-path",
				},
			}))
		})

		Context("when stemcell file does not exist", func() {
			BeforeEach(func() {
				fakeFs.RemoveAll("fake-stemcell-path")
			})

			It("returns error", func() {
				_, err := validator.ValidateStemcell("fake-stemcell-path", fakeStemcellExtractor)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Verifying that the stemcell `fake-stemcell-path' exists"))

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Validating stemcell",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Verifying that the stemcell `fake-stemcell-path' exists",
				}))
			})
		})
	})
})
