package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/gofrs/uuid"
	. "github.com/kai5263499/rhema"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"

	"github.com/sirupsen/logrus"
)

var (
	cfg *domain.Config
)

func main() {
	cfg = &domain.Config{}
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

	scrape := NewScrape(cfg)

	for _, arg := range os.Args[1:] {
		newUUID := uuid.Must(uuid.NewV4())

		req := &pb.Request{
			Title:       newUUID.String(),
			Type:        pb.ContentType_URI,
			Created:     uint64(time.Now().Unix()),
			Uri:         arg,
			RequestHash: newUUID.String(),
		}

		scrapeConvertErr := scrape.Convert(req)
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
