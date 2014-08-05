package release_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/release"
	faketar "github.com/cloudfoundry/bosh-micro-cli/tar/fakes"
)

var _ = Describe("tarReader", func() {
	var (
		reader        Reader
		tarFilePath   string
		fakeFs        *fakesys.FakeFileSystem
		fakeExtractor *faketar.FakeExtractor
	)

	BeforeEach(func() {
		tarFilePath = "/some/release.tgz"
		fakeFs = fakesys.NewFakeFileSystem()
		fakeExtractor = faketar.NewFakeExtractor()
		reader = NewTarReader(tarFilePath, fakeFs, fakeExtractor)
	})

	Describe("Read", func() {
		Context("when the given release is readable", func() {
			BeforeEach(func() {
				fakeExtractor.ExtractPath = "/somedir"
				fakeFs.WriteFileString(
					"/somedir/release.MF",
					`---
name: fake-release
version: fake-version

commit_hash: abc123
uncommitted_changes: true

jobs:
- name: fake-job
  version: fake-job-version
  fingerprint: fake-job-fingerprint
  sha1: fake-job-sha

packages:
- name: fake-package
  version: fake-package-version
  fingerprint: fake-package-fingerprint
  sha1: fake-package-sha
  dependencies:
  - fake-package-1
`,
				)
			})

			It("returns a release from the given tar file", func() {
				release, err := reader.Read()
				Expect(err).NotTo(HaveOccurred())
				Expect(release.Name).To(Equal("fake-release"))
				Expect(release.Version).To(Equal("fake-version"))
				Expect(release.CommitHash).To(Equal("abc123"))
				Expect(release.UncommittedChanges).To(BeTrue())

				Expect(len(release.Jobs)).To(Equal(1))
				Expect(release.Jobs).To(
					ContainElement(
						Job{
							Name:        "fake-job",
							Version:     "fake-job-version",
							Fingerprint: "fake-job-fingerprint",
							Sha1:        "fake-job-sha",
						},
					),
				)

				Expect(len(release.Packages)).To(Equal(1))
				Expect(release.Packages).To(
					ContainElement(
						Package{
							Name:         "fake-package",
							Version:      "fake-package-version",
							Fingerprint:  "fake-package-fingerprint",
							Sha1:         "fake-package-sha",
							Dependencies: []string{"fake-package-1"},
						},
					),
				)
			})

			It("reads binary values for version, fingerprint, and sha1 properties of jobs and packages", func() {
				fakeFs.WriteFileString(
					"/somedir/release.MF",
					`---
jobs:
- version: !binary |-
    ZGQxYmEzMzBiYzQ0YjMxODFiMjYzMzgzYjhlNDI1MmQ3MDUxZGVjYQ==
  fingerprint: !binary |-
    ZGQxYmEzMzBiYzQ0YjMxODFiMjYzMzgzYjhlNDI1MmQ3MDUxZGVjYQ==
  sha1: !binary |-
    NmVhYTZjOTYxZWFjN2JkOTk0ZDE2NDRhZDQwNWIzMzk1NDIwZWNhZg==

packages:
- version: !binary |-
    ZGQxYmEzMzBiYzQ0YjMxODFiMjYzMzgzYjhlNDI1MmQ3MDUxZGVjYQ==
  fingerprint: !binary |-
    ZGQxYmEzMzBiYzQ0YjMxODFiMjYzMzgzYjhlNDI1MmQ3MDUxZGVjYQ==
  sha1: !binary |-
    NmVhYTZjOTYxZWFjN2JkOTk0ZDE2NDRhZDQwNWIzMzk1NDIwZWNhZg==
`,
				)

				release, err := reader.Read()
				Expect(err).NotTo(HaveOccurred())
				Expect(len(release.Jobs)).To(Equal(1))
				Expect(release.Jobs).To(
					ContainElement(
						Job{
							Version:     "dd1ba330bc44b3181b263383b8e4252d7051deca",
							Fingerprint: "dd1ba330bc44b3181b263383b8e4252d7051deca",
							Sha1:        "6eaa6c961eac7bd994d1644ad405b3395420ecaf",
						},
					),
				)

				Expect(len(release.Packages)).To(Equal(1))
				Expect(release.Packages).To(
					ContainElement(
						Package{
							Version:     "dd1ba330bc44b3181b263383b8e4252d7051deca",
							Fingerprint: "dd1ba330bc44b3181b263383b8e4252d7051deca",
							Sha1:        "6eaa6c961eac7bd994d1644ad405b3395420ecaf",
						},
					),
				)
			})

			It("returns all errors when the binary values are invalid", func() {
				fakeFs.WriteFileString(
					"/somedir/release.MF",
					`---
jobs:
- name: fake-job-1
  version: !binary |-
    fake-binary==
  fingerprint: !binary |-
    fake-binary==
  sha1: !binary |-
    fake-binary==
- name: fake-job-2
  version: !binary |-
    fake-binary==
  fingerprint: !binary |-
    fake-binary==
  sha1: !binary |-
    fake-binary==

packages:
- name: fake-package-1
  version: !binary |-
    fake-binary==
  fingerprint: !binary |-
    fake-binary==
  sha1: !binary |-
    fake-binary==
- name: fake-package-2
  version: !binary |-
    fake-binary==
  fingerprint: !binary |-
    fake-binary==
  sha1: !binary |-
    fake-binary==
`,
				)

				_, err := reader.Read()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Decoding binary job version for 'fake-job-1'"))
				Expect(err.Error()).To(ContainSubstring("Decoding binary job fingerprint for 'fake-job-1'"))
				Expect(err.Error()).To(ContainSubstring("Decoding binary job sha for 'fake-job-1'"))
				Expect(err.Error()).To(ContainSubstring("Decoding binary job version for 'fake-job-2'"))
				Expect(err.Error()).To(ContainSubstring("Decoding binary job fingerprint for 'fake-job-2'"))
				Expect(err.Error()).To(ContainSubstring("Decoding binary job sha for 'fake-job-2'"))
				Expect(err.Error()).To(ContainSubstring("Decoding binary package version for 'fake-package-1'"))
				Expect(err.Error()).To(ContainSubstring("Decoding binary package fingerprint for 'fake-package-1'"))
				Expect(err.Error()).To(ContainSubstring("Decoding binary package sha for 'fake-package-1'"))
				Expect(err.Error()).To(ContainSubstring("Decoding binary package version for 'fake-package-2'"))
				Expect(err.Error()).To(ContainSubstring("Decoding binary package fingerprint for 'fake-package-2'"))
				Expect(err.Error()).To(ContainSubstring("Decoding binary package sha for 'fake-package-2'"))
			})
		})

		Context("when the CPI release manifest is invalid YAML", func() {
			BeforeEach(func() {
				fakeExtractor.ExtractPath = "/somedir"
				fakeFs.WriteFileString("/somedir/release.MF", "{")
			})

			It("return err", func() {
				_, err := reader.Read()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Parsing release manifest"))
			})
		})

		Context("when the CPI release does not have a release manifest", func() {
			It("return err", func() {
				_, err := reader.Read()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Reading release manifest"))
			})
		})

		Context("when the CPI release is not a valid tar", func() {
			BeforeEach(func() {
				fakeExtractor.ExtractError = errors.New("")
			})

			It("returns err", func() {
				_, err := reader.Read()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Extracting release"))
			})
		})
	})

	Describe("Close", func() {
		BeforeEach(func() {
			fakeExtractor.ExtractPath = "/somedir"
			fakeFs.WriteFileString("/somedir/release.MF", "{}")
		})

		Context("when release is extracted", func() {
			It("cleans up the extracted release", func() {
				_, err := reader.Read()
				Expect(err).ToNot(HaveOccurred())
				err = reader.Close()
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeFs.FileExists("/somedir/release.MF")).NotTo(BeTrue())
			})
		})

		Context("when release is not yet extracted", func() {
			It("does nothing", func() {
				err := reader.Close()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
