package rhema_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRhema(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rhema Suite")
}
