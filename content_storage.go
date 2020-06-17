package rhema

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"

	"github.com/kai5263499/rhema/interfaces"
)

const (
	esIndex   = "requests"
	esMapping = `
{
	"mappings":{
		"properties":{
			"Title":{
				"type":"keyword"
			},
			"Text":{
				"type":"text"
			},
			"Created":{
				"type":"date"
			},
			"Type":{
				"type":"integer"
			},
			"Size":{
				"type":"integer"
			},
			"Length":{
				"type":"integer"
			},
			"SubmittedAt":{
				"type":"date"
			},
			"SubmittedBy":{
				"type":"keyword"
			},
			"RequestHash":{
				"type": "keyword"
			},
			"Uri":{
				"type": "keyword"
			},
			"DownloadURI":{
				"type": "keyword"
			}
		}
	}
}`
)

// NewContentStorage returns an instance of ContentStorage
func NewContentStorage(
	tmpPath string,
	bucket string,
	gcpClient interfaces.GCPStorage,
	copyToLocal bool,
	localPath string,
	chownTo int,
	copyToCloud bool,
) (*ContentStorage, error) {

	cs := &ContentStorage{
		tmpPath:       tmpPath,
		localPath:     localPath,
		copyToLocal:   copyToLocal,
		bucket:        bucket,
		storageClient: gcpClient,
		chownTo:       chownTo,
		copyToCloud:   copyToCloud,
	}

	return cs, nil
}

// ContentStorage persists content artifacts to S3
type ContentStorage struct {
	localPath     string
	tmpPath       string
	copyToLocal   bool
	bucket        string
	chownTo       int
	copyToCloud   bool
	storageClient interfaces.GCPStorage
}

func copyFileContents(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	err = out.Sync()
	return nil
}

// Store persists a content item in S3
func (cs *ContentStorage) Store(ci pb.Request) (pb.Request, error) {
	itemPath, err := getPath(ci)
	if err != nil {
		return ci, err
	}

	if cs.copyToLocal && ci.Type == pb.Request_AUDIO {
		src := filepath.Join(cs.tmpPath, itemPath)
		dst := filepath.Join(cs.localPath, filepath.Base(itemPath))

		if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
			logrus.WithError(err).Errorf("unable to MkdirAll on %s", dst)
			return ci, err
		}

		if err := copyFileContents(src, dst); err != nil {
			logrus.WithError(err).Errorf("unable to copy %s to %s", src, dst)
			return ci, err
		}

		if err := os.Chown(dst, cs.chownTo, cs.chownTo); err != nil {
			logrus.WithError(err).Errorf("unable to chown %s to %d", dst, cs.chownTo)
			return ci, err
		}

		logrus.Debugf("%s -> %s", src, dst)
	}

	if cs.copyToCloud {
		if err := cs.doCloudStore(&ci, itemPath); err != nil {
			return ci, err
		}

		if err := cs.presign(&ci); err != nil {
			return ci, err
		}
	}

	return ci, nil
}

func (cs *ContentStorage) presign(ci *pb.Request) error {
	itemPath, err := getPath(*ci)
	if err != nil {
		return err
	}

	jsonKey, err := ioutil.ReadFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if err != nil {
		return fmt.Errorf("cannot read the JSON key file, err: %v", err)
	}

	conf, err := google.JWTConfigFromJSON(jsonKey)
	if err != nil {
		return fmt.Errorf("google.JWTConfigFromJSON: %v", err)
	}

	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         "GET",
		GoogleAccessID: conf.Email,
		PrivateKey:     conf.PrivateKey,
		Expires:        time.Now().Add(24 * time.Hour),
	}

	urlStr, err := storage.SignedURL(cs.bucket, itemPath, opts)
	if err != nil {
		return fmt.Errorf("Unable to generate a signed URL: %v", err)
	}

	if err != nil {
		logrus.WithError(err).Errorf("unable to get object presigned uri")
		return err
	}

	ci.DownloadURI = urlStr

	logrus.Debugf("downloadURI=%s", ci.DownloadURI)

	return nil
}

func (cs *ContentStorage) doCloudStore(ci *pb.Request, path string) error {
	fileName, err := GetFilePath(*ci)
	if err != nil {
		return nil
	}

	fullFileName := filepath.Join(cs.tmpPath, fileName)

	// Open the file for use
	f, err := os.Open(fullFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := f.Stat()
	var size int64 = fileInfo.Size()

	logrus.Debugf("storing %d bytes from %s to %s/%s", size, fullFileName, cs.bucket, path)

	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	wc := cs.storageClient.Bucket(cs.bucket).Object(path).NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		logrus.WithError(err).Errorf("unable to copy file to bucket")
		return err
	}
	if err := wc.Close(); err != nil {
		logrus.WithError(err).Errorf("unable to close bucket writer")
		return err
	}

	if err != nil {
		logrus.WithError(err).Errorf("unable to put object into S3")
		return err
	}

	return nil
}

func (cs *ContentStorage) SetConfig(key string, value string) bool {
	switch key {
	case "bucket":
		cs.bucket = value
		return true
	case "localpath":
		cs.localPath = value
		return true
	}

	return false
}

func (cs *ContentStorage) GetConfig(key string) (bool, string) {
	switch key {
	case "bucket":
		return true, cs.bucket
	case "localpath":
		return true, cs.localPath
	default:
		return false, ""
	}
}
