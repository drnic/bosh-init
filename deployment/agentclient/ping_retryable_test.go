package agentclient_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshretry "github.com/cloudfoundry/bosh-agent/retrystrategy"
	fakebmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/agentclient"
)

var _ = Describe("PingRetryable", func() {
	Describe("Attempt", func() {
		var (
			fakeAgentClient *fakebmagentclient.FakeAgentClient
			pingRetryable   boshretry.Retryable
		)

		BeforeEach(func() {
			fakeAgentClient = fakebmagentclient.NewFakeAgentClient()
			pingRetryable = NewPingRetryable(fakeAgentClient)
		})

		It("tells the agent client to ping", func() {
			isRetryable, err := pingRetryable.Attempt()
			Expect(err).ToNot(HaveOccurred())
			Expect(isRetryable).To(BeTrue())
			Expect(fakeAgentClient.PingCalledCount).To(Equal(1))
		})

		Context("when pinging fails", func() {
			BeforeEach(func() {
				fakeAgentClient.SetPingBehavior("", errors.New("fake-agent-client-ping-error"))
			})

			It("returns an error", func() {
				isRetryable, err := pingRetryable.Attempt()
				Expect(err).To(HaveOccurred())
				Expect(isRetryable).To(BeTrue())
				Expect(err.Error()).To(ContainSubstring("fake-agent-client-ping-error"))
			})
		})
	})
})
