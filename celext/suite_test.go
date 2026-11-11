package celext_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCelext(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "celext")
}
