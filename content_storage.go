package rhema

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"time"

	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// NewContentStorage returns an instance of ContentStorage
func NewContentStorage(s3svc s3iface.S3API, localPath string, s3Bucket string) *ContentStorage {
	return &ContentStorage{
		s3svc:     s3svc,
		localPath: localPath,
		s3Bucket:  s3Bucket,
	}
}

// ContentStorage persists content artifacts to S3
type ContentStorage struct {
	s3svc     s3iface.S3API
	localPath string
	s3Bucket  string
}

// Store persists a content item in S3
func (cs *ContentStorage) Store(ci pb.Request) (pb.Request, error) {
	var err error

	fileName, err := GetFilePath(ci)
	fullFileName := filepath.Join(cs.localPath, fileName)

	// Open the file for use
	file, err := os.Open(fullFileName)
	if err != nil {
		return ci, err
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	_, err = file.Read(buffer)
	if err != nil {
		return ci, err
	}

	s3path, err := getS3Path(ci)
	if err != nil {
		return ci, err
	}

	logrus.Debugf("storing %d bytes from %s to %s", size, fullFileName, s3path)

	_, err = cs.s3svc.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(cs.s3Bucket),
		Key:                  aws.String(s3path),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})

	if err != nil {
		return ci, err
	}

	req, _ := cs.s3svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(cs.s3Bucket),
		Key:    aws.String(s3path),
	})
	urlStr, err := req.Presign(6 * 24 * time.Hour)

	if err != nil {
		return ci, err
	}

	ci.DownloadURI = urlStr

	logrus.Debugf("downloadURI=%s\n", ci.DownloadURI)

	return ci, nil
}
