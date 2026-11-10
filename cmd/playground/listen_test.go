package main

import (
	"net"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPlaygroundListener(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Playground listener")
}

var _ = Describe("evaluation listener allocation", func() {
	It("retains the selected port while serving and uses another port on conflict", func() {
		occupied, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(occupied.Close)
		listener, err := listen(occupied.Addr().String(), true)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(listener.Close)
		Expect(listener.Addr().String()).NotTo(Equal(occupied.Addr().String()))
		competitor, err := net.Listen("tcp", listener.Addr().String())
		if competitor != nil {
			competitor.Close()
		}
		Expect(err).To(HaveOccurred())
	})
	It("reports occupied ports when dynamic allocation is disabled", func() {
		occupied, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(occupied.Close)
		_, err = listen(occupied.Addr().String(), false)
		Expect(err).To(HaveOccurred())
	})
	It("does not hide invalid address errors", func() {
		_, err := listen("127.0.0.1:invalid", true)
		Expect(err).To(HaveOccurred())
	})
})
