package rhema

import (
	"io/ioutil"
	"os"
	"path"

	"cloud.google.com/go/storage"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"

	"path/filepath"

	"github.com/gofrs/uuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type placeholderGCPClient struct{}

func (c *placeholderGCPClient) Bucket(name string) *storage.BucketHandle {
	return &storage.BucketHandle{}
}

var _ = Describe("content_storage", func() {
	PIt("Should store the text file in GCP", func() {
		var err error

		cfg := &domain.Config{}

		cs, err := NewContentStorage(cfg, nil)
		Expect(err).To(BeNil())

		requestContent := "this is the scraped text data from a url request"

		newUUID := uuid.Must(uuid.NewV4())

		ci := &pb.Request{
			Created: 383576400,
			Type:    pb.ContentType_TEXT,
			Title:   newUUID.String(),
			Length:  uint64(len(requestContent)),
			Text:    requestContent,
		}

		fileName, err := GetFilePath(ci)
		Expect(err).To(BeNil())

		txtFilename := filepath.Join(cs.cfg.LocalPath, fileName)

		err = os.MkdirAll(path.Dir(txtFilename), os.ModePerm)
		Expect(err).To(BeNil())

		err = ioutil.WriteFile(txtFilename, []byte(ci.Text), 0644)
		Expect(err).To(BeNil())

		err = cs.Store(ci)
		Expect(err).To(BeNil())

		_, err = getPath(ci)
		Expect(err).To(BeNil())

		err = os.Remove(txtFilename)
		Expect(err).To(BeNil())
	})
})
