package rhema

import (
	"io/ioutil"
	"os"
	"path"
	"time"

	pb "github.com/kai5263499/rhema/generated"

	"path/filepath"

	"github.com/gofrs/uuid"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var lastPutInput *s3.PutObjectInput

type mockS3Client struct {
	s3iface.S3API
}

func (mc *mockS3Client) PutObject(p *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	lastPutInput = p
	return nil, nil
}

func (mc *mockS3Client) GetObjectRequest(*s3.GetObjectInput) (*request.Request, *s3.GetObjectOutput) {
	req := &request.Request{}
	return req, nil
}

type presignerMock struct{}

func (p *presignerMock) Presign(d time.Duration) (string, error) {
	return "yay", nil
}

var _ = Describe("content_storage", func() {
	PIt("Should store the text file in S3", func() {
		var err error

		mockClient := mockS3Client{}

		cs := NewContentStorage(&mockClient, "/tmp", "my-bucket")

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

		s3Key, err := getS3Path(ci)
		Expect(err).To(BeNil())

		Expect(lastPutInput).To(Not(BeNil()))
		Expect(*lastPutInput.Key).To(Equal(s3Key))

		err = os.Remove(txtFilename)
		Expect(err).To(BeNil())
	})
})
