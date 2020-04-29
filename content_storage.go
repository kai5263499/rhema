package rhema

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"time"

	pb "github.com/kai5263499/rhema/generated"
	"github.com/olivere/elastic"
	"github.com/sirupsen/logrus"

	"github.com/kai5263499/rhema/interfaces"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
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
func NewContentStorage(s3svc interfaces.S3, localPath string, s3Bucket string, esClient interfaces.ElasticSearch) *ContentStorage {

	cs := &ContentStorage{
		s3svc:     s3svc,
		localPath: localPath,
		s3Bucket:  s3Bucket,
		esClient:  esClient,
	}

	esSetup(esClient)

	return cs
}

// ContentStorage persists content artifacts to S3
type ContentStorage struct {
	s3svc     interfaces.S3
	localPath string
	s3Bucket  string
	esClient  interfaces.ElasticSearch
}

func esSetup(esClient interfaces.ElasticSearch) {
	ctx := context.Background()

	exists, err := esClient.IndexExists(esIndex).Do(ctx)
	if err != nil {
		logrus.WithError(err).Fatalf("index existance check")
	}
	if !exists {
		logrus.Debugf("index doesn't exist, creating")
		createIndex, err := esClient.CreateIndex(esIndex).BodyString(esMapping).Do(ctx)
		if err != nil {
			logrus.WithError(err).Fatalf("index creation not acknowledged")
		}
		if !createIndex.Acknowledged {
			logrus.Warnf("index creation not acknowledged")
		}
		logrus.Debugf("index created")
	}
}

// Store persists a content item in S3
func (cs *ContentStorage) Store(ci pb.Request) (pb.Request, error) {
	var err error

	s3path, err := getS3Path(ci)
	if err != nil {
		return ci, err
	}

	err = cs.s3store(&ci, s3path)
	if err != nil {
		return ci, err
	}

	req, _ := cs.s3svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(cs.s3Bucket),
		Key:    aws.String(s3path),
	})

	err = cs.presignS3(&ci, req)
	if err != nil {
		return ci, err
	}

	err = cs.esStore(&ci)
	if err != nil {
		logrus.WithError(err).Errorf("esStore failed")
		return ci, err
	}

	return ci, nil
}

func (cs *ContentStorage) presignS3(ci *pb.Request, req interfaces.Request) error {
	urlStr, err := req.Presign(6 * 24 * time.Hour)

	if err != nil {
		logrus.WithError(err).Errorf("unable to get object presigned uri")
		return err
	}

	ci.DownloadURI = urlStr

	logrus.Debugf("downloadURI=%s", ci.DownloadURI)

	return nil
}

func (cs *ContentStorage) s3store(ci *pb.Request, s3path string) error {
	fileName, err := GetFilePath(*ci)
	if err != nil {
		return nil
	}

	fullFileName := filepath.Join(cs.localPath, fileName)

	// Open the file for use
	file, err := os.Open(fullFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	logrus.Debugf("storing %d bytes from %s to s3://%s/%s", size, fullFileName, cs.s3Bucket, s3path)

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
		logrus.WithError(err).Errorf("unable to put object into S3")
		return err
	}

	return nil
}

func (cs *ContentStorage) esStore(ci *pb.Request) error {
	ctx := context.Background()

	termQuery := elastic.NewTermQuery("RequestHash", ci.RequestHash)
	searchResult, err := cs.esClient.Search().
		Index(esIndex).
		Query(termQuery).
		From(0).Size(1).
		Do(ctx)
	if err != nil {
		return err
	}

	if searchResult.TotalHits() > 0 {
		hit := searchResult.Hits.Hits[0]

		logrus.Debugf("got document %s in version %d from index %s, type %s\n", hit.Id, hit.Version, hit.Index, hit.Type)

		resp, err := cs.esClient.Update().
			Index(esIndex).
			Id(hit.Id).
			DocAsUpsert(true).
			Doc(ci).
			Do(ctx)
		if err != nil {
			return err
		}

		logrus.Debugf("es upserted doc status=%s", resp.Status)
	} else {

		logrus.WithFields(logrus.Fields{
			"RequestHash": ci.RequestHash,
		}).Debugf("es index new doc")

		resp, err := cs.esClient.Index().
			Index(esIndex).
			BodyJson(ci).
			Do(ctx)
		if err != nil {
			return err
		}

		logrus.Debugf("es index new doc status=%s", resp.Status)
	}

	return nil
}

func (cs *ContentStorage) SetConfig(key string, value string) bool {
	switch key {
	case "s3bucket":
		cs.s3Bucket = value
		return true
	case "localpath":
		cs.localPath = value
		return true
	}

	return false
}

func (cs *ContentStorage) GetConfig(key string) (bool, string) {
	switch key {
	case "s3bucket":
		return true, cs.s3Bucket
	case "localpath":
		return true, cs.localPath
	default:
		return false, ""
	}
}
