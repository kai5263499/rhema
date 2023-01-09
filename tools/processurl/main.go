package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/kai5263499/rhema/client"
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

	processorClient, err := client.NewClientWithResponses(cfg.RequestProcessorUri)
	if err != nil {
		logrus.WithError(err).Error("error creating processor client")
		return
	}

	sri := []client.SubmitRequestInput{}

	for _, arg := range os.Args[1:] {
		ri := client.SubmitRequestInput{
			Uri:            arg,
			Atempo:         &cfg.Atempo,
			WordsPerMinute: &cfg.WordsPerMinute,
			EspeakVoice:    &cfg.EspeakVoice,
			SubmittedWith:  &cfg.SubmittedWith,
		}

		contentType := pb.ContentType_URI.String()
		ri.Type = &contentType

		now := uint64(time.Now().UTC().Unix())
		intnow := int(now)
		ri.SubmittedAt = &intnow

		sri = append(sri, ri)
	}

	logrus.Debugf("submitting %d requests to %s", len(sri), cfg.RequestProcessorUri)

	resp, err := processorClient.SubmitRequestWithResponse(context.Background(), &client.SubmitRequestParams{}, sri)
	if err != nil {
		logrus.WithError(err).Error("error sending request")
		return
	}

	if resp.StatusCode() != http.StatusAccepted {
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode(),
			"body":        string(resp.Body),
		}).Error("error with request response")
		return
	}

	logrus.Infof("successfully submitted %d requests", len(sri))
}
