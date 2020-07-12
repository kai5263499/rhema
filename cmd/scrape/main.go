package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env"
	"github.com/gofrs/uuid"
	. "github.com/kai5263499/rhema"
	pb "github.com/kai5263499/rhema/generated"

	"github.com/sirupsen/logrus"
)

type config struct {
	AwsDefaultRegion string `env:"AWS_DEFAULT_REGION" envDefault:"us-east-1"`
	MinTextBlockSize int    `env:"MIN_TEXT_BLOCK_SIZE" envDefault:"100"`
	LocalPath        string `env:"LOCAL_PATH" envDefault:"/tmp"`
	LogLevel         string `env:"LOG_LEVEL" envDefault:"info"`
	TitleLengthLimit int    `env:"TITLE_LENGTH_LIMIT" envDefault:"40"`
}

var (
	cfg config
)

func main() {
	cfg = config{}
	if err := env.Parse(&cfg); err != nil {
		logrus.WithError(err).Fatal("parse config")
	}

	if len(os.Args) < 2 {
		logrus.Errorf("wrong number of arguments\n")
	}

	if level, err := logrus.ParseLevel(cfg.LogLevel); err != nil {
		logrus.WithError(err).Fatal("parse level")
	} else {
		logrus.SetLevel(level)
	}

	scrape := NewScrape(uint32(cfg.MinTextBlockSize), cfg.LocalPath, cfg.TitleLengthLimit)

	for _, arg := range os.Args[1:] {
		newUUID := uuid.Must(uuid.NewV4())

		req := pb.Request{
			Title:       newUUID.String(),
			Type:        pb.ContentType_URI,
			Created:     uint64(time.Now().Unix()),
			Uri:         arg,
			RequestHash: newUUID.String(),
		}

		var scrapeConvertErr error
		req, scrapeConvertErr = scrape.Convert(req)
		if scrapeConvertErr != nil {
			logrus.WithError(scrapeConvertErr).Error("scrape convert")
			continue
		}

		bytes, jsonMarshalErr := json.Marshal(req)
		if jsonMarshalErr != nil {
			logrus.WithError(jsonMarshalErr).Error("json marshal")
			continue
		}

		fmt.Println(string(bytes))
	}
}
