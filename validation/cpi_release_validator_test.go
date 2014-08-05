package validation_test

import (
	"errors"
	"os"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakes "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/cloudfoundry/bosh-micro-cli/validation"
)

var _ = Describe("CPIReleaseValidator", func() {
	var (
		extractedReleasePath string
		fakeFs               *fakes.FakeFileSystem
		validator            CPIReleaseValidator
	)

	BeforeEach(func() {
		fakeFs = fakes.NewFakeFileSystem()
		extractedReleasePath, _ = fakeFs.TempDir("validation-cpiReleaseValidator")
		validator = NewCPIReleaseValidator(fakeFs)
	})

	Describe("Validate", func() {
		Context("when release is valid", func() {
			BeforeEach(func() {
				cpiJobPath := path.Join(extractedReleasePath, "jobs", "cpi")
				fakeFs.MkdirAll(path.Join(cpiJobPath, "templates"), os.ModePerm)
				fakeFs.WriteFileString(path.Join(cpiJobPath, "spec"), "---\ntemplates:\n  cpi_script: bin/cpi")
				fakeFs.WriteFileString(path.Join(cpiJobPath, "templates", "cpi_script"), "")
			})

			It("does not return an error", func() {
				err := validator.Validate(extractedReleasePath)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("when the release is invalid", func() {
			Context("when the job spec exists", func() {
				var cpiJobPath string
				BeforeEach(func() {
					cpiJobPath = path.Join(extractedReleasePath, "jobs", "cpi")
					fakeFs.MkdirAll(path.Join(cpiJobPath, "templates"), os.ModePerm)
					fakeFs.WriteFileString(path.Join(cpiJobPath, "spec"), "")
				})

				Context("when the spec cannot be parsed", func() {
					BeforeEach(func() {
						fakeFs.WriteFileString(path.Join(cpiJobPath, "spec"), "{")
					})

					It("should error", func() {
						err := validator.Validate(extractedReleasePath)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Unmarshalling job spec"))
					})
				})

				Context("when the bin/cpi template does not exist", func() {
					BeforeEach(func() {
						fakeFs.WriteFileString(path.Join(cpiJobPath, "spec"), "---\ntemplates:\n  cpi_script: bin/cpi")
					})

					It("should error", func() {
						err := validator.Validate(extractedReleasePath)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("'bin/cpi' template file 'cpi_script' does not exist in job"))
					})
				})

				Context("when the spec does not contain bin/cpi", func() {
					BeforeEach(func() {
						fakeFs.WriteFileString(path.Join(cpiJobPath, "spec"), "---\n{}")
					})

					It("should error", func() {
						err := validator.Validate(extractedReleasePath)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("No template exists that will render 'bin/cpi' in job spec"))
					})
				})
			})

			Context("when the spec does not exist", func() {
				BeforeEach(func() {
					fakeFs.ReadFileError = errors.New("")
				})

				It("should error", func() {
					err := validator.Validate(extractedReleasePath)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Reading job spec"))
				})
			})
		})
	})
})
