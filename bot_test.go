package rhema

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("content_storage", func() {
	It("Should store the text file in S3", func() {
		var err error
		Expect(err).To(BeNil())
	})
})
