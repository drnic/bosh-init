package tar_test

import (
	"errors"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/tar"
)

var _ = Describe("cmdExtractor", func() {
	var (
		fakeRunner *fakesys.FakeCmdRunner
		fakeFs     *fakesys.FakeFileSystem
		logger     boshlog.Logger
		extractor  Extractor
	)

	BeforeEach(func() {
		fakeRunner = fakesys.NewFakeCmdRunner()
		fakeFs = fakesys.NewFakeFileSystem()
		fakeFs.TempDirDir = "/some/tmpdir"
		logger = boshlog.NewLogger(boshlog.LevelNone)
		extractor = NewCmdExtractor(fakeRunner, fakeFs, logger)
	})

	Describe("Extract", func() {
		It("extracts the tar into a new tempdir and returns the extracted path", func() {
			path, err := extractor.Extract("/some/tar")
			Expect(err).NotTo(HaveOccurred())
			Expect(path).To(Equal("/some/tmpdir"))
			Expect(fakeRunner.RunCommands).To(ContainElement([]string{"tar", "-C", "/some/tmpdir", "-xzf", "/some/tar"}))
		})

		Context("when the tempdir cannot be created", func() {
			BeforeEach(func() {
				fakeFs.TempDirError = errors.New("")
			})

			It("returns err", func() {
				_, err := extractor.Extract("/some/tar")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Creating tempdir for tar extraction"))
			})
		})

		Context("when the tar cannot be extracted", func() {
			BeforeEach(func() {
				result := fakesys.FakeCmdResult{
					Error: errors.New(""),
				}
				fakeRunner.AddCmdResult("tar -C /some/tmpdir -xzf /some/tar", result)
			})

			It("returns err", func() {
				_, err := extractor.Extract("/some/tar")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Extracting tar '/some/tar'"))
			})

			It("cleans up the extracted path", func() {
				_, _ = extractor.Extract("/some/tar")
				Expect(fakeFs.FileExists("/some/tmpdir")).To(Equal(false))
			})
		})
	})
})
