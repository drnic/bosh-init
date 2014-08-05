package cmd_test

import (
	"fmt"
	"io/ioutil"
	"os"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmcmd "github.com/cloudfoundry/bosh-micro-cli/cmd"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmtar "github.com/cloudfoundry/bosh-micro-cli/tar"
	bmtestutils "github.com/cloudfoundry/bosh-micro-cli/testutils"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("DeploymentCmd", func() {
	var (
		command    bmcmd.Cmd
		config     bmconfig.Config
		fakeFs     *fakesys.FakeFileSystem
		fakeUI     *fakeui.FakeUI
		fakeRunner *fakesys.FakeCmdRunner
		extractor  bmtar.Extractor
	)

	BeforeEach(func() {
		fakeUI = &fakeui.FakeUI{}
		fakeFs = fakesys.NewFakeFileSystem()
		config = bmconfig.Config{}
		fakeRunner = fakesys.NewFakeCmdRunner()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		extractor = bmtar.NewCmdExtractor(fakeRunner, fakeFs, logger)

		command = bmcmd.NewDeployCmd(fakeUI, config, fakeFs, extractor)
	})

	Describe("Run", func() {
		Context("when there is a deployment set", func() {
			BeforeEach(func() {
				config.Deployment = "/some/deployment/file"
				command = bmcmd.NewDeployCmd(fakeUI, config, fakeFs, extractor)
			})

			Context("when the deployment manifest exists", func() {
				BeforeEach(func() {
					fakeFs.WriteFileString(config.Deployment, "")
				})

				Context("when no arguments are given", func() {
					It("returns err", func() {
						err := command.Run([]string{})
						Expect(err).To(HaveOccurred())
						Expect(fakeUI.Errors).To(ContainElement("No CPI release provided"))
					})
				})

				Context("when a CPI release is given", func() {
					Context("and the CPI release is valid", func() {
						BeforeEach(func() {
							// TODO: make a valid CPI release
							fakeFs.WriteFileString("/somepath", "")
						})

						It("does not return an error", func() {
							err := command.Run([]string{"/somepath"})
							Expect(err).NotTo(HaveOccurred())
						})
					})

					Context("and the CPI release is invalid", func() {
						var (
							tarFilePath string
							tempFile    *os.File
						)

						BeforeEach(func() {
							var err error
							tempFile, err = ioutil.TempFile("", "deployCmdTest")
							Expect(err).NotTo(HaveOccurred())
							fakeFs.ReturnTempFile = tempFile
							tarFilePath, err = bmtestutils.GenerateTarfile(fakeFs, []bmtestutils.TarFileContent{})
							Expect(err).NotTo(HaveOccurred())
						})
						AfterEach(func() {
							err := os.RemoveAll(tempFile.Name())
							Expect(err).NotTo(HaveOccurred())
						})

						It("returns err", func() {
							err := command.Run([]string{tarFilePath})
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("Validating CPI release"))
							Expect(fakeUI.Errors).To(ContainElement(fmt.Sprintf("CPI release '%s' is not a valid BOSH release:", tarFilePath)))
						})
					})

					Context("and the CPI release does not exist", func() {
						It("returns err", func() {
							err := command.Run([]string{"/somepath"})
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("Validating CPI release"))
							Expect(fakeUI.Errors).To(ContainElement("CPI release '/somepath' does not exist"))
						})
					})
				})
			})

			Context("when the deployment manifest is missing", func() {
				BeforeEach(func() {
					config.Deployment = "/some/deployment/file"
					command = bmcmd.NewDeployCmd(fakeUI, config, fakeFs, extractor)
				})

				It("returns err", func() {
					err := command.Run([]string{"/somepath"})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Reading deployment manifest for deploy"))
					Expect(fakeUI.Errors).To(ContainElement("Deployment manifest path '/some/deployment/file' does not exist"))
				})
			})
		})

		Context("when there is no deployment set", func() {
			It("returns err", func() {
				err := command.Run([]string{"/somepath"})
				Expect(err).To(HaveOccurred())
				Expect(fakeUI.Errors).To(ContainElement("No deployment set"))
			})
		})
	})
})
