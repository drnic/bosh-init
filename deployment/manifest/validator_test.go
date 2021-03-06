package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrelmanifest "github.com/cloudfoundry/bosh-micro-cli/release/manifest"
	bmrelset "github.com/cloudfoundry/bosh-micro-cli/release/set"

	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
)

var _ = Describe("Validator", func() {
	var (
		logger         boshlog.Logger
		releaseManager bmrel.Manager
		validator      Validator

		releases      []bmrelmanifest.ReleaseRef
		validManifest Manifest
		fakeRelease   *fakebmrel.FakeRelease
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
		releaseManager = bmrel.NewManager(logger)

		releases = []bmrelmanifest.ReleaseRef{
			{
				Name:    "fake-release-name",
				Version: "1.0",
			},
		}

		validManifest = Manifest{
			Name: "fake-deployment-name",
			Networks: []Network{
				{
					Name: "fake-network-name",
					Type: "dynamic",
				},
			},
			ResourcePools: []ResourcePool{
				{
					Name:    "fake-resource-pool-name",
					Network: "fake-network-name",
					RawCloudProperties: map[interface{}]interface{}{
						"fake-prop-key": "fake-prop-value",
						"fake-prop-map-key": map[interface{}]interface{}{
							"fake-prop-key": "fake-prop-value",
						},
					},
				},
			},
			DiskPools: []DiskPool{
				{
					Name:     "fake-disk-pool-name",
					DiskSize: 1024,
					RawCloudProperties: map[interface{}]interface{}{
						"fake-prop-key": "fake-prop-value",
						"fake-prop-map-key": map[interface{}]interface{}{
							"fake-prop-key": "fake-prop-value",
						},
					},
				},
			},
			Jobs: []Job{
				{
					Name: "fake-job-name",
					Templates: []ReleaseJobRef{
						{
							Name:    "fake-job-name",
							Release: "fake-release-name",
						},
					},
					PersistentDisk: 1024,
					Networks: []JobNetwork{
						{
							Name:      "fake-network-name",
							StaticIPs: []string{"127.0.0.1"},
							Default:   []NetworkDefault{"dns", "gateway"},
						},
					},
					Lifecycle: "service",
					RawProperties: map[interface{}]interface{}{
						"fake-prop-key": "fake-prop-value",
						"fake-prop-map-key": map[interface{}]interface{}{
							"fake-prop-key": "fake-prop-value",
						},
					},
				},
			},
			RawProperties: map[interface{}]interface{}{
				"fake-prop-key": "fake-prop-value",
				"fake-prop-map-key": map[interface{}]interface{}{
					"fake-prop-key": "fake-prop-value",
				},
			},
		}

		fakeRelease = fakebmrel.New("fake-release-name", "1.0")
		fakeRelease.ReleaseJobs = []bmrel.Job{{Name: "fake-job-name"}}
		releaseManager.Add(fakeRelease)
	})

	JustBeforeEach(func() {
		releaseResolver := bmrelset.NewResolver(releaseManager, logger)
		err := releaseResolver.Filter(releases)
		Expect(err).ToNot(HaveOccurred())
		validator = NewValidator(logger, releaseResolver)
	})

	Describe("Validate", func() {
		It("does not error if deployment is valid", func() {
			deploymentManifest := validManifest

			err := validator.Validate(deploymentManifest)
			Expect(err).ToNot(HaveOccurred())
		})

		It("validates name is not empty", func() {
			deploymentManifest := Manifest{
				Name: "",
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("name must be provided"))
		})

		It("validates name is not blank", func() {
			deploymentManifest := Manifest{
				Name: "   \t",
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("name must be provided"))
		})

		It("validates that there is only one resource pool", func() {
			deploymentManifest := Manifest{
				ResourcePools: []ResourcePool{
					{},
					{},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools must be of size 1"))
		})

		It("validates resource pool name", func() {
			deploymentManifest := Manifest{
				ResourcePools: []ResourcePool{
					{
						Name: "",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[0].name must be provided"))

			deploymentManifest = Manifest{
				ResourcePools: []ResourcePool{
					{
						Name: "not-blank",
					},
					{
						Name: "   \t",
					},
				},
			}

			err = validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[1].name must be provided"))
		})

		It("validates resource pool network", func() {
			deploymentManifest := Manifest{
				ResourcePools: []ResourcePool{
					{
						Network: "",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[0].network must be provided"))

			deploymentManifest = Manifest{
				Networks: []Network{
					{
						Name: "fake-network",
					},
				},
				ResourcePools: []ResourcePool{
					{
						Network: "other-network",
					},
				},
			}

			err = validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[0].network must be the name of a network"))
		})

		It("validates resource pool cloud_properties", func() {
			deploymentManifest := Manifest{
				ResourcePools: []ResourcePool{
					{
						RawCloudProperties: map[interface{}]interface{}{
							123: "fake-property-value",
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[0].cloud_properties must have only string keys"))
		})

		It("validates resource pool env", func() {
			deploymentManifest := Manifest{
				ResourcePools: []ResourcePool{
					{
						RawEnv: map[interface{}]interface{}{
							123: "fake-env-value",
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("resource_pools[0].env must have only string keys"))
		})

		It("validates disk pool name", func() {
			deploymentManifest := Manifest{
				DiskPools: []DiskPool{
					{
						Name: "",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("disk_pools[0].name must be provided"))

			deploymentManifest = Manifest{
				DiskPools: []DiskPool{
					{
						Name: "not-blank",
					},
					{
						Name: "   \t",
					},
				},
			}

			err = validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("disk_pools[1].name must be provided"))
		})

		It("validates disk pool size", func() {
			deploymentManifest := Manifest{
				DiskPools: []DiskPool{
					{
						Name: "fake-disk",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("disk_pools[0].disk_size must be > 0"))
		})

		It("validates disk pool cloud_properties", func() {
			deploymentManifest := Manifest{
				DiskPools: []DiskPool{
					{
						RawCloudProperties: map[interface{}]interface{}{
							123: "fake-property-value",
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("disk_pools[0].cloud_properties must have only string keys"))
		})

		It("validates network name", func() {
			deploymentManifest := Manifest{
				Networks: []Network{
					{
						Name: "",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("networks[0].name must be provided"))

			deploymentManifest = Manifest{
				Networks: []Network{
					{
						Name: "not-blank",
					},
					{
						Name: "   \t",
					},
				},
			}

			err = validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("networks[1].name must be provided"))
		})

		It("validates network type", func() {
			deploymentManifest := Manifest{
				Networks: []Network{
					{
						Type: "unknown-type",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("networks[0].type must be 'manual', 'dynamic', or 'vip'"))
		})

		It("validates disk pool cloud_properties", func() {
			deploymentManifest := Manifest{
				Networks: []Network{
					{
						RawCloudProperties: map[interface{}]interface{}{
							123: "fake-property-value",
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("networks[0].cloud_properties must have only string keys"))
		})

		It("validates that there is only one job", func() {
			deploymentManifest := Manifest{
				Jobs: []Job{
					{},
					{},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs must be of size 1"))
		})

		It("validates job name", func() {
			deploymentManifest := Manifest{
				Jobs: []Job{
					{
						Name: "",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].name must be provided"))

			deploymentManifest = Manifest{
				Jobs: []Job{
					{
						Name: "not-blank",
					},
					{
						Name: "   \t",
					},
				},
			}

			err = validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[1].name must be provided"))
		})

		It("validates job persistent_disk", func() {
			deploymentManifest := Manifest{
				Jobs: []Job{
					{
						PersistentDisk: -1234,
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].persistent_disk must be >= 0"))
		})

		It("validates job persistent_disk_pool", func() {
			deploymentManifest := Manifest{
				Jobs: []Job{
					{
						PersistentDiskPool: "non-existent-disk-pool",
					},
				},
				DiskPools: []DiskPool{
					{
						Name: "fake-disk-pool",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].persistent_disk_pool must be the name of a disk pool"))
		})

		It("validates job instances", func() {
			deploymentManifest := Manifest{
				Jobs: []Job{
					{
						Instances: -1234,
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].instances must be >= 0"))
		})

		It("validates job networks", func() {
			deploymentManifest := Manifest{
				Jobs: []Job{
					{
						Networks: []JobNetwork{},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].networks must be a non-empty array"))
		})

		It("validates job network name", func() {
			deploymentManifest := Manifest{
				Jobs: []Job{
					{
						Networks: []JobNetwork{
							{
								Name: "",
							},
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].networks[0].name must be provided"))
		})

		It("validates job network static ips", func() {
			deploymentManifest := Manifest{
				Jobs: []Job{
					{
						Networks: []JobNetwork{
							{
								StaticIPs: []string{"non-ip"},
							},
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].networks[0].static_ips[0] must be a valid IP"))
		})

		It("validates job network default", func() {
			deploymentManifest := Manifest{
				Jobs: []Job{
					{
						Networks: []JobNetwork{
							{
								Default: []NetworkDefault{
									"non-dns-or-gateway",
								},
							},
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].networks[0].default[0] must be 'dns' or 'gateway'"))
		})

		It("validates job lifecycle", func() {
			deploymentManifest := Manifest{
				Jobs: []Job{
					{
						Lifecycle: "errand",
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].lifecycle must be 'service' ('errand' not supported)"))
		})

		It("validates job properties", func() {
			deploymentManifest := Manifest{
				Jobs: []Job{
					{
						RawProperties: map[interface{}]interface{}{
							123: "fake-property-value",
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].properties must have only string keys"))
		})

		It("permits job templates to reference an undeclared release", func() {
			deploymentManifest := validManifest
			deploymentManifest.Jobs[0].Templates = []ReleaseJobRef{
				{
					Name:    "fake-job-name",
					Release: "fake-release-name",
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).NotTo(HaveOccurred())
		})

		It("validates job templates have a job name", func() {
			deploymentManifest := validManifest
			deploymentManifest.Jobs = []Job{
				{
					Templates: []ReleaseJobRef{{}},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].templates[0].name must be provided"))
		})

		It("validates job templates have unique job names", func() {
			deploymentManifest := validManifest
			deploymentManifest.Jobs = []Job{
				{
					Templates: []ReleaseJobRef{
						{Name: "fake-job-name"},
						{Name: "fake-job-name"},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].templates[1].name 'fake-job-name' must be unique"))
		})

		It("validates job templates reference a release", func() {
			deploymentManifest := Manifest{
				Jobs: []Job{
					{
						Templates: []ReleaseJobRef{
							{Name: "fake-job-name"},
						},
					},
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].templates[0].release must be provided"))
		})

		It("validates job templates reference an available release", func() {
			deploymentManifest := validManifest
			deploymentManifest.Jobs[0].Templates = []ReleaseJobRef{
				{Name: "fake-job-name", Release: "fake-other-release-name"},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].templates[0].release must refer to an available release"))
		})

		It("validates job templates reference an available release", func() {
			deploymentManifest := validManifest
			deploymentManifest.Jobs[0].Templates = []ReleaseJobRef{
				{Name: "fake-job-name", Release: "fake-other-release-name"},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].templates[0].release must refer to an available release"))
		})

		It("validates job templates reference a job declared within the release", func() {
			deploymentManifest := validManifest
			deploymentManifest.Jobs[0].Templates = []ReleaseJobRef{
				{Name: "fake-other-job-name", Release: "fake-release-name"},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("jobs[0].templates[0] must refer to a job in 'fake-release-name', but there is no job named 'fake-other-job-name'"))
		})

		It("validates deployment properties", func() {
			deploymentManifest := Manifest{
				RawProperties: map[interface{}]interface{}{
					123: "fake-property-value",
				},
			}

			err := validator.Validate(deploymentManifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("properties must have only string keys"))
		})
	})
})
