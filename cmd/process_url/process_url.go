package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/caarlos0/env"
	"github.com/gofrs/uuid"
	. "github.com/kai5263499/rhema"
	. "github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
)

type config struct {
	AwsDefaultRegion string  `env:"AWS_DEFAULT_REGION" envDefault:"us-east-1"`
	MinTextBlockSize int     `env:"MIN_TEXT_BLOCK_SIZE" envDefault:"100"`
	S3Bucket         string  `env:"S3_BUCKET"`
	TmpPath          string  `env:"TMP_PATH" envDefault:"/tmp"`
	WordsPerMinute   int     `env:"WORDS_PER_MINUTE" envDefault:"250"`
	EspeakVoice      string  `env:"ESPEAK_VOICE" envDefault:"f5"`
	LocalPath        string  `env:"LOCAL_PATH" envDefault:"/data"`
	Atempo           float32 `env:"ATEMPO" envDefault:"2.0"`
	ChownTo          int     `env:"CHOWN_TO" envDefault:"1000"`
	LogLevel         string  `env:"LOG_LEVEL" envDefault:"info"`
}

var (
	cfg config
)

func main() {
	var err error
	cfg = config{}
	err = env.Parse(&cfg)
	CheckError(err)

	logrus.ParseLevel(cfg.LogLevel)

	s3svc := s3.New(session.New(aws.NewConfig().WithRegion(cfg.AwsDefaultRegion).WithCredentials(credentials.NewEnvCredentials())))

	contentStorage := NewContentStorage(s3svc, cfg.TmpPath, cfg.S3Bucket)

	speedupAudo := NewSpeedupAudio(contentStorage, cfg.TmpPath, cfg.Atempo)

	scrape := NewScrape(contentStorage, uint32(cfg.MinTextBlockSize), cfg.TmpPath)
	youtube := NewYoutube(scrape, contentStorage, speedupAudo, cfg.TmpPath)
	text2mp3 := NewText2Mp3(contentStorage, cfg.TmpPath, cfg.WordsPerMinute, cfg.EspeakVoice)
	requestProcessor := NewRequestProcessor(cfg.TmpPath, scrape, youtube, text2mp3, speedupAudo)

	for _, arg := range os.Args[1:] {
		newUUID := uuid.Must(uuid.NewV4())

		req := pb.Request{
			Title:       newUUID.String(),
			Type:        pb.Request_URI,
			Created:     uint64(time.Now().Unix()),
			Uri:         arg,
			RequestHash: newUUID.String(),
		}

		resultingItem, err := requestProcessor.Process(req)
		CheckError(err)

		urlFilename, err := GetFilePath(resultingItem)
		CheckError(err)

		baseUrlFilename := path.Base(urlFilename)

		urlFullFilename := filepath.Join(cfg.LocalPath, baseUrlFilename)

		err = DownloadUriToFile(resultingItem.DownloadURI, urlFullFilename)
		CheckError(err)

		// Get file info
		file, err := os.Open(urlFullFilename)
		CheckError(err)
		fileInfo, err := file.Stat()
		CheckError(err)
		var size int64 = fileInfo.Size()
		file.Close()

		err = os.Chown(urlFullFilename, cfg.ChownTo, cfg.ChownTo)
		CheckError(err)

		fmt.Printf("%s\t%d\n", urlFullFilename, size)
	}
}
