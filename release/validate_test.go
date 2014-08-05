package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/release"
)

var _ = Describe("Validate", func() {
	Context("given a valid release", func() {
		var release Release

		BeforeEach(func() {
			release = Release{
				Name:               "fake-release-name",
				Version:            "fake-release-version",
				CommitHash:         "fake-release-commit-hash",
				UncommittedChanges: true,

				Jobs: []Job{
					{
						Name:        "fake-job-1-name",
						Version:     "fake-job-1-version",
						Fingerprint: "fake-job-1-fingerprint",
						Sha1:        "fake-job-1-sha",
					},
				},

				Packages: []Package{
					{
						Name:        "fake-package-1-name",
						Version:     "fake-package-1-version",
						Fingerprint: "fake-package-1-fingerprint",
						Sha1:        "fake-package-1-sha",
						Dependencies: []string{
							"fake-package-1-dependency-1",
							"fake-package-1-dependency-2",
						},
					},
				},
			}
		})

		It("validates that release without error", func() {
			err := Validate(release)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("given a invalid release", func() {
		// It("fails due to lack")
	})
})
