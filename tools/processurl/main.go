package main

import (
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/gofrs/uuid"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
)

var (
	cfg *domain.Config
)

func main() {
	cfg = &domain.Config{}
	if err := env.Parse(cfg); err != nil {
		logrus.WithError(err).Fatal("parse config")
	}

	if level, err := logrus.ParseLevel(cfg.LogLevel); err != nil {
		logrus.WithError(err).Fatal("parse log level")
	} else {
		logrus.SetLevel(level)
	}

	for _, arg := range os.Args[1:] {
		newUUID := uuid.Must(uuid.NewV4())

		req := &pb.Request{
			Title:          newUUID.String(),
			Type:           pb.ContentType_URI,
			Created:        uint64(time.Now().Unix()),
			Uri:            arg,
			RequestHash:    newUUID.String(),
			SubmittedBy:    cfg.SubmittedWith,
			SubmittedAt:    uint64(time.Now().Unix()),
			ATempo:         cfg.Atempo,
			WordsPerMinute: cfg.WordsPerMinute,
			ESpeakVoice:    cfg.EspeakVoice,
			SubmittedWith:  cfg.SubmittedWith,
		}

		logrus.Infof("request successfully submitted to processor %+#v", req)
	}
}
