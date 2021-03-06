package pkg_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmpkgs "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakebminstallblob "github.com/cloudfoundry/bosh-micro-cli/installation/blob/fakes"
	fakebmpkgs "github.com/cloudfoundry/bosh-micro-cli/installation/pkg/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
)

var _ = Describe("Install", func() {
	var (
		installer     PackageInstaller
		blobExtractor *fakebminstallblob.FakeExtractor
		repo          *fakebmpkgs.FakeCompiledPackageRepo
		targetDir     string
		pkg           *bmrel.Package
	)
	BeforeEach(func() {
		repo = fakebmpkgs.NewFakeCompiledPackageRepo()
		blobExtractor = fakebminstallblob.NewFakeExtractor()
		targetDir = "fake-target-dir"
		installer = NewPackageInstaller(repo, blobExtractor)

		pkg = &bmrel.Package{
			Name:         "fake-package-name",
			Fingerprint:  "fake-package-fingerprint",
			SHA1:         "fake-package-sha1",
			Dependencies: []*bmrel.Package{},
		}
	})

	Context("when the package exists in the repo", func() {
		BeforeEach(func() {
			record := bmpkgs.CompiledPackageRecord{
				BlobID:   "fake-blob-id",
				BlobSHA1: "fake-package-fingerprint",
			}
			repo.SetFindBehavior(*pkg, record, true, nil)
			blobExtractor.SetExtractBehavior("fake-blob-id", "fake-package-fingerprint", "fake-target-dir/fake-package-name", nil)
		})

		It("gets the package record from the repo", func() {
			err := installer.Install(pkg, targetDir)
			Expect(err).ToNot(HaveOccurred())
		})

		It("extracts the blob into the target dir", func() {
			err := installer.Install(pkg, targetDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(blobExtractor.ExtractInputs).To(ContainElement(fakebminstallblob.ExtractInput{
				BlobID:    "fake-blob-id",
				BlobSHA1:  "fake-package-fingerprint",
				TargetDir: "fake-target-dir/fake-package-name",
			}))
		})

		Context("when finding the package in the repo errors", func() {
			BeforeEach(func() {
				repo.SetFindBehavior(*pkg, bmpkgs.CompiledPackageRecord{}, false, errors.New("fake-error"))
			})

			It("returns an error", func() {
				err := installer.Install(pkg, targetDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Finding compiled package record"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})
	})

	Context("when the package does not exist in the repo", func() {
		BeforeEach(func() {
			repo.SetFindBehavior(*pkg, bmpkgs.CompiledPackageRecord{}, false, nil)
		})

		It("returns an error", func() {
			err := installer.Install(pkg, targetDir)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Compiled package record not found"))
		})
	})
})
