package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/olivere/elastic/v7"

	"github.com/caarlos0/env"
	"github.com/gofrs/uuid"
	. "github.com/kai5263499/rhema"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
)

type config struct {
	MinTextBlockSize             int     `env:"MIN_TEXT_BLOCK_SIZE" envDefault:"100"`
	Bucket                       string  `env:"BUCKET"`
	TmpPath                      string  `env:"TMP_PATH" envDefault:"/tmp"`
	WordsPerMinute               int     `env:"WORDS_PER_MINUTE" envDefault:"250"`
	EspeakVoice                  string  `env:"ESPEAK_VOICE" envDefault:"f5"`
	LocalPath                    string  `env:"LOCAL_PATH" envDefault:"/data"`
	Atempo                       float32 `env:"ATEMPO" envDefault:"2.0"`
	ChownTo                      int     `env:"CHOWN_TO" envDefault:"1000"`
	LogLevel                     string  `env:"LOG_LEVEL" envDefault:"info"`
	TitleLengthLimit             int     `env:"TITLE_LENGTH_LIMIT" envDefault:"40"`
	ElasticSearchAddress         string  `env:"ELASTICSEARCH_URL" envDefault:"http://localhost:9200"`
	GoogleApplicationCredentials string  `env:"GOOGLE_APPLICATION_CREDENTIALS"`
}

var (
	cfg config
)

func main() {
	cfg = config{}
	if err := env.Parse(&cfg); err != nil {
		logrus.WithError(err).Fatal("parse config")
	}

	if level, err := logrus.ParseLevel(cfg.LogLevel); err != nil {
		logrus.WithError(err).Fatal("parse log level")
	} else {
		logrus.SetLevel(level)
	}

	esClient, newESClientErr := elastic.NewClient(elastic.SetURL(cfg.ElasticSearchAddress))
	if newESClientErr != nil {
		logrus.WithError(newESClientErr).Fatal("new elasticsearch client")
	}

	ctx := context.Background()
	gcpClient, newGCPClientErr := storage.NewClient(ctx)
	if newGCPClientErr != nil {
		logrus.WithError(newGCPClientErr).Fatal("new gcp storage client")
	}

	contentStorage, newContentStorageErr := NewContentStorage(cfg.TmpPath, cfg.Bucket, gcpClient, esClient)
	if newContentStorageErr != nil {
		logrus.WithError(newContentStorageErr).Fatal("new content storage")
	}

	speedupAudo := NewSpeedupAudio(contentStorage, cfg.TmpPath, cfg.Atempo)

	scrape := NewScrape(contentStorage, uint32(cfg.MinTextBlockSize), cfg.TmpPath, cfg.TitleLengthLimit)
	youtube := NewYoutube(scrape, contentStorage, speedupAudo, cfg.TmpPath)
	text2mp3 := NewText2Mp3(contentStorage, cfg.TmpPath, cfg.WordsPerMinute, cfg.EspeakVoice)
	requestProcessor := NewRequestProcessor(cfg.TmpPath, scrape, youtube, text2mp3, speedupAudo, cfg.TitleLengthLimit)

	for _, arg := range os.Args[1:] {
		newUUID := uuid.Must(uuid.NewV4())

		req := pb.Request{
			Title:       newUUID.String(),
			Type:        pb.Request_URI,
			Created:     uint64(time.Now().Unix()),
			Uri:         arg,
			RequestHash: newUUID.String(),
		}

		resultingItem, processRequestErr := requestProcessor.Process(req)
		if processRequestErr != nil {
			logrus.WithError(processRequestErr).Error("process request error")
			continue
		}

		urlFilename, getFilePathErr := GetFilePath(resultingItem)
		if getFilePathErr != nil {
			logrus.WithError(processRequestErr).Error("get file path")
			continue
		}

		baseUrlFilename := path.Base(urlFilename)

		urlFullFilename := filepath.Join(cfg.LocalPath, baseUrlFilename)

		if err := DownloadUriToFile(resultingItem.DownloadURI, urlFullFilename); err != nil {
			logrus.WithError(err).Error("process request error")
			continue
		}

		// Get file info
		file, fileOpenErr := os.Open(urlFullFilename)
		if fileOpenErr != nil {
			logrus.WithError(fileOpenErr).Errorf("error opening %s", urlFullFilename)
			continue
		}

		fileInfo, fileStatErr := file.Stat()
		if fileStatErr != nil {
			logrus.WithError(fileStatErr).Error("file stat")
			continue
		}

		var size int64 = fileInfo.Size()
		file.Close()

		if err := os.Chown(urlFullFilename, cfg.ChownTo, cfg.ChownTo); err != nil {
			logrus.WithError(err).Error("chown")
			continue
		}

		fmt.Printf("%s\t%d\n", urlFullFilename, size)
	}
}
