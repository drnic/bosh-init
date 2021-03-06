package erbrenderer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/erbrenderer"
)

var _ = Describe("PropertiesResolver", func() {
	var (
		propertiesResolver PropertiesResolver
		defaults           map[string]interface{}
		values             map[string]interface{}
	)

	Context("when value is specified for nested property", func() {
		BeforeEach(func() {
			values = map[string]interface{}{
				"first-level-prop": map[string]interface{}{
					"second-level-prop": "original-value",
				},
			}
			defaults = map[string]interface{}{
				"first-level-prop.second-level-prop": "default-value",
			}

			propertiesResolver = NewPropertiesResolver(defaults, values)
		})

		It("returns the specified value", func() {
			properties := propertiesResolver.Resolve()
			Expect(properties).To(Equal(map[string]interface{}{
				"first-level-prop": map[string]interface{}{
					"second-level-prop": "original-value",
				},
			}))
		})
	})

	Context("when value is not specified for nested property", func() {
		BeforeEach(func() {
			values = map[string]interface{}{}
		})

		Context("when default property is specified", func() {
			BeforeEach(func() {
				defaults = map[string]interface{}{
					"first-level-prop.second-level-prop": "default-value",
				}

				propertiesResolver = NewPropertiesResolver(defaults, values)
			})

			It("uses default property", func() {
				properties := propertiesResolver.Resolve()
				Expect(properties).To(Equal(map[string]interface{}{
					"first-level-prop": map[string]interface{}{
						"second-level-prop": "default-value",
					},
				}))
			})
		})

		Context("when default property is nil", func() {
			BeforeEach(func() {
				defaults = map[string]interface{}{
					"first-level-prop.second-level-prop": nil,
				}

				propertiesResolver = NewPropertiesResolver(defaults, values)
			})

			It("uses default property", func() {
				properties := propertiesResolver.Resolve()
				Expect(properties).To(Equal(map[string]interface{}{
					"first-level-prop": map[string]interface{}{
						"second-level-prop": nil,
					},
				}))
			})
		})

		Context("when default property is empty string", func() {
			BeforeEach(func() {
				defaults = map[string]interface{}{
					"first-level-prop.second-level-prop": "",
				}

				propertiesResolver = NewPropertiesResolver(defaults, values)
			})

			It("uses default property", func() {
				properties := propertiesResolver.Resolve()
				Expect(properties).To(Equal(map[string]interface{}{
					"first-level-prop": map[string]interface{}{
						"second-level-prop": "",
					},
				}))
			})
		})
	})
})
