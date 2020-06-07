package rhema

import (
	"io/ioutil"
	"os"
	"path"

	pb "github.com/kai5263499/rhema/generated"

	"path/filepath"

	"github.com/gofrs/uuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("content_storage", func() {
	PIt("Should store the text file in S3", func() {
		var err error

		cs := NewContentStorage("/tmp", "my-bucket")

		requestContent := "this is the scraped text data from a url request"

		newUUID := uuid.Must(uuid.NewV4())

		ci := pb.Request{
			Created: 383576400,
			Type:    pb.Request_TEXT,
			Title:   newUUID.String(),
			Length:  uint64(len(requestContent)),
			Text:    requestContent,
		}

		fileName, err := GetFilePath(ci)
		Expect(err).To(BeNil())

		txtFilename := filepath.Join(cs.localPath, fileName)

		err = os.MkdirAll(path.Dir(txtFilename), os.ModePerm)
		Expect(err).To(BeNil())

		err = ioutil.WriteFile(txtFilename, []byte(ci.Text), 0644)
		Expect(err).To(BeNil())

		_, err = cs.Store(ci)
		Expect(err).To(BeNil())

		_, err = getPath(ci)
		Expect(err).To(BeNil())

		err = os.Remove(txtFilename)
		Expect(err).To(BeNil())
	})
})
