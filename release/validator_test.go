package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/release"
)

var _ = Describe("Validator", func() {
	var fakeFs *fakesys.FakeFileSystem

	BeforeEach(func() {
		fakeFs = fakesys.NewFakeFileSystem()
	})

	It("validates a valid release without error", func() {
		fakeFs.WriteFileString("/some/job/path/monit", "")
		fakeFs.WriteFileString("/some/job/path/templates/fake-job-1-template", "")
		release := NewRelease(
			"fake-release-name",
			"fake-release-version",
			[]Job{
				{
					Name:          "fake-job-1-name",
					Fingerprint:   "fake-job-1-fingerprint",
					SHA1:          "fake-job-1-sha",
					Templates:     map[string]string{"fake-job-1-template": "fake-job-1-file"},
					ExtractedPath: "/some/job/path",
				},
			},
			[]*Package{
				{
					Name:        "fake-package-1-name",
					Fingerprint: "fake-package-1-fingerprint",
					SHA1:        "fake-package-1-sha",
					Dependencies: []*Package{
						&Package{Name: "fake-package-1-dependency-1"},
						&Package{Name: "fake-package-1-dependency-2"},
					},
				},
			},
			"/some/release/path",
			fakeFs,
		)
		validator := NewValidator(fakeFs)

		err := validator.Validate(release)
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns all errors with an empty release", func() {
		validator := NewValidator(fakeFs)
		release := NewRelease(
			"",
			"",
			[]Job{},
			[]*Package{},
			"",
			fakeFs,
		)
		err := validator.Validate(release)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Release name is missing"))
		Expect(err.Error()).To(ContainSubstring("Release version is missing"))
	})

	It("returns all errors with jobs and packages in a release", func() {
		release := NewRelease(
			"fake-release-name",
			"fake-release-version",
			[]Job{{}, {Name: "fake-job"}},
			[]*Package{{}, {Name: "fake-package"}},
			"/some/release/path",
			fakeFs,
		)
		validator := NewValidator(fakeFs)

		err := validator.Validate(release)
		Expect(err).To(HaveOccurred())

		Expect(err.Error()).To(ContainSubstring("Job name is missing"))
		Expect(err.Error()).To(ContainSubstring("Job '' fingerprint is missing"))
		Expect(err.Error()).To(ContainSubstring("Job '' sha1 is missing"))
		Expect(err.Error()).To(ContainSubstring("Job 'fake-job' fingerprint is missing"))
		Expect(err.Error()).To(ContainSubstring("Job 'fake-job' sha1 is missing"))

		Expect(err.Error()).To(ContainSubstring("Package name is missing"))
		Expect(err.Error()).To(ContainSubstring("Package '' fingerprint is missing"))
		Expect(err.Error()).To(ContainSubstring("Package '' sha1 is missing"))
		Expect(err.Error()).To(ContainSubstring("Package 'fake-package' fingerprint is missing"))
		Expect(err.Error()).To(ContainSubstring("Package 'fake-package' sha1 is missing"))
	})

	Context("when jobs are missing templates", func() {
		It("returns errors with each job that is missing templates", func() {
			release := NewRelease(
				"fake-release",
				"fake-version",
				[]Job{
					{
						Name:        "fake-job",
						Fingerprint: "fake-fingerprint",
						SHA1:        "fake-sha",
						Templates:   map[string]string{"fake-template": "fake-file"},
					},
					{
						Name:        "fake-job-2",
						Fingerprint: "fake-fingerprint-2",
						SHA1:        "fake-sha-2",
						Templates:   map[string]string{"fake-template-2": "fake-file-2"},
					},
				},
				[]*Package{},
				"/some/release/path",
				fakeFs,
			)
			validator := NewValidator(fakeFs)

			err := validator.Validate(release)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("Job 'fake-job' is missing template 'templates/fake-template'"))
			Expect(err.Error()).To(ContainSubstring("Job 'fake-job-2' is missing template 'templates/fake-template-2'"))
		})
	})

	Context("when jobs are missing monit", func() {
		It("returns erros with each job that is missing monit", func() {
			release := NewRelease(
				"fake-release-name",
				"fake-release-version",
				[]Job{
					{
						Name:        "fake-job-1",
						Fingerprint: "fake-finger-print-1",
						SHA1:        "fake-sha-1",
					},
					{
						Name:        "fake-job-2",
						Fingerprint: "fake-finger-print-2",
						SHA1:        "fake-sha-2",
					},
				},
				[]*Package{{}, {Name: "fake-package"}},
				"/some/release/path",
				fakeFs,
			)
			validator := NewValidator(fakeFs)

			err := validator.Validate(release)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("Job 'fake-job-1' is missing monit file"))
			Expect(err.Error()).To(ContainSubstring("Job 'fake-job-2' is missing monit file"))
		})
	})

	Context("when jobs have package dependencies that are not in the release", func() {
		It("returns errors with each job that is missing packages", func() {
			release := NewRelease(
				"fake-release-name",
				"fake-release-version",
				[]Job{
					{
						Name:         "fake-job",
						Fingerprint:  "fake-fingerprint",
						SHA1:         "fake-sha",
						Templates:    map[string]string{},
						PackageNames: []string{"fake-package"},
					},
					{
						Name:         "fake-job-2",
						Fingerprint:  "fake-fingerprint-2",
						SHA1:         "fake-sha-2",
						Templates:    map[string]string{},
						PackageNames: []string{"fake-package-2"},
					},
				},
				[]*Package{},
				"/some/release/path",
				fakeFs,
			)
			validator := NewValidator(fakeFs)

			err := validator.Validate(release)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("Job 'fake-job' requires 'fake-package' which is not in the release"))
			Expect(err.Error()).To(ContainSubstring("Job 'fake-job-2' requires 'fake-package-2' which is not in the release"))
		})
	})
})
