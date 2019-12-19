package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env"
	"github.com/gofrs/uuid"
	. "github.com/kai5263499/rhema"
	. "github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"

	"github.com/sirupsen/logrus"
)

type config struct {
	AwsDefaultRegion string `env:"AWS_DEFAULT_REGION" envDefault:"us-east-1"`
	MinTextBlockSize int    `env:"MIN_TEXT_BLOCK_SIZE" envDefault:"100"`
	LocalPath        string `env:"LOCAL_PATH" envDefault:"/tmp"`
	LogLevel         string `env:"LOG_LEVEL" envDefault:"info"`
}

var (
	cfg config
)

type placeholderContentStore struct{}

func (fs *placeholderContentStore) Store(ci pb.Request) (pb.Request, error) { return ci, nil }

func main() {
	var err error
	cfg = config{}
	err = env.Parse(&cfg)
	CheckError(err)

	if len(os.Args) < 2 {
		logrus.Errorf("wrong number of arguments\n")
	}

	logrus.ParseLevel(cfg.LogLevel)

	scrape := NewScrape(&placeholderContentStore{}, uint32(cfg.MinTextBlockSize), cfg.LocalPath)

	for _, arg := range os.Args[1:] {
		newUUID := uuid.Must(uuid.NewV4())

		req := pb.Request{
			Title:       newUUID.String(),
			Type:        pb.Request_URI,
			Created:     uint64(time.Now().Unix()),
			Uri:         arg,
			RequestHash: newUUID.String(),
		}

		req, err = scrape.Convert(req)
		CheckError(err)

		bytes, err := json.Marshal(req)
		CheckError(err)

		fmt.Println(string(bytes))
	}
}
