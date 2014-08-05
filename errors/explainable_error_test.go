package errors_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/errors"
)

var _ = Describe("explainableError", func() {
	Describe("Error", func() {
		Context("when reasons are given", func() {
			It("returns the header and each reason as bullet points", func() {
				err := NewExplainableError("My reasons:")
				err.AddError(errors.New("reason 1"))
				err.AddError(errors.New("reason 2"))
				Expect(err.Error()).To(Equal("My reasons:\n* reason 1\n* reason 2"))
			})
		})

		Context("when no header and some reasons are given", func() {
			It("returns the reasons with no header", func() {
				err := NewExplainableError("")
				err.AddError(errors.New("reason 1"))
				err.AddError(errors.New("reason 2"))
				Expect(err.Error()).To(Equal("* reason 1\n* reason 2"))
			})
		})

		Context("when no reasons are given", func() {
			It("returns the header and no bullet points", func() {
				err := NewExplainableError("My reasons:")
				Expect(err.Error()).To(Equal("My reasons:"))
			})
		})

		Context("when no header and no reasons are given", func() {
			It("returns empty string", func() {
				err := NewExplainableError("")
				Expect(err.Error()).To(Equal(""))
			})
		})
	})

	Describe("HasErrors", func() {
		Context("when reasons are given", func() {
			It("returns true", func() {
				err := NewExplainableError("")
				err.AddError(errors.New("reason"))
				Expect(err.HasErrors()).To(BeTrue())
			})
		})

		Context("when no reasons are given", func() {
			It("returns false", func() {
				err := NewExplainableError("")
				Expect(err.HasErrors()).To(BeFalse())
			})
		})
	})
})
