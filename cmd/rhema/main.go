package main

import (
	"context"
	"flag"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	. "github.com/kai5263499/rhema"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
)

var (
	interactive      bool
	requestProcessor *RequestProcessor
	speedupAudo      *SpeedupAudio
	scrape           *Scrape
	youtube          *YouTube
	text2mp3         *Text2Mp3
	contentStorage   domain.Storage
	cfg              = &domain.Config{}
)

func main() {
	if err := env.Parse(cfg); err != nil {
		logrus.WithError(err).Fatal("parse config")
	}

	if level, err := logrus.ParseLevel(cfg.LogLevel); err != nil {
		logrus.WithError(err).Fatal("parse log level")
	} else {
		logrus.SetLevel(level)
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(true)

	flag.BoolVar(&interactive, "i", false, "Interactive")
	flag.Parse()

	speedupAudo = NewSpeedupAudio(cfg, exec.Command)
	scrape = NewScrape(cfg)
	text2mp3 = NewText2Mp3(cfg)
	youtube = NewYoutube(cfg, scrape, speedupAudo)

	var err error
	contentStorage, err = NewContentStorage(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("new storage client")
	}

	logrus.Debug("content storage created")

	requestProcessor = NewRequestProcessor(cfg,
		scrape,
		youtube,
		text2mp3,
		speedupAudo,
		contentStorage,
	)

	if !interactive {
		startApiServer()
	} else {
		startInteractive()
	}
}

func startApiServer() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	api, err := NewApi(
		ctx,
		stop,
		cfg,
		requestProcessor,
		contentStorage,
	)
	if err != nil {
		logrus.WithError(err).Fatal("new api")
	}

	api.Start()

	<-ctx.Done()

	logrus.Info("api exiting")
}

func startInteractive() {
	logrus.Debug("running interactively")

	sri := []*pb.Request{}

	now := uint64(time.Now().UTC().Unix())
	intnow := uint64(now)

	for _, arg := range os.Args[1:] {
		ri := &pb.Request{
			Uri:            arg,
			ATempo:         cfg.Atempo,
			WordsPerMinute: cfg.WordsPerMinute,
			ESpeakVoice:    cfg.EspeakVoice,
			SubmittedWith:  cfg.SubmittedWith,
			Type:           pb.ContentType_URI,
			SubmittedAt:    intnow,
		}

		sri = append(sri, ri)
	}

	logrus.Debugf("parsed %d urls", len(sri))

	var wg sync.WaitGroup
	for _, req := range sri {
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			if err := requestProcessor.Process(req); err != nil {
				logrus.WithError(err).Errorf("error processing request for uri %s", req.Uri)
			}
		}(&wg)
		wg.Add(1)
	}

	wg.Wait()
}
