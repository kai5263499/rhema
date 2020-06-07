package main

import (
	"context"
	"os"
	"os/signal"

	"cloud.google.com/go/storage"
	"github.com/caarlos0/env"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"

	. "github.com/kai5263499/rhema"
)

type config struct {
	SlackToken                   string   `env:"SLACK_TOKEN"`
	MinTextBlockSize             int      `env:"MIN_TEXT_BLOCK_SIZE" envDefault:"100"`
	Bucket                       string   `env:"BUCKET"`
	TmpPath                      string   `env:"TMP_PATH" envDefault:"/tmp"`
	WordsPerMinute               int      `env:"WORDS_PER_MINUTE" envDefault:"350"`
	EspeakVoice                  string   `env:"ESPEAK_VOICE" envDefault:"f5"`
	LocalPath                    string   `env:"LOCAL_PATH" envDefault:"/data"`
	Atempo                       float32  `env:"ATEMPO" envDefault:"2.0"`
	ChownTo                      int      `env:"CHOWN_TO" envDefault:"1000"`
	LogLevel                     string   `env:"LOG_LEVEL" envDefault:"info"`
	Channels                     []string `env:"CHANNELS" envDefault:"content"`
	TitleLengthLimit             int      `env:"TITLE_LENGTH_LIMIT" envDefault:"40"`
	ElasticSearchAddress         string   `env:"ELASTICSEARCH_URL" envDefault:"http://localhost:9200"`
	GoogleApplicationCredentials string   `env:"GOOGLE_APPLICATION_CREDENTIALS"`
}

var (
	cfg config
)

func main() {
	cfg = config{}
	if err := env.Parse(&cfg); err != nil {
		logrus.WithError(err).Fatal("parse configs")
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
	gcpClient, newGCPStorageErr := storage.NewClient(ctx)
	if newGCPStorageErr != nil {
		logrus.WithError(newGCPStorageErr).Fatal("new gcp storage client")
	}

	contentStorage, newStorageErr := NewContentStorage(cfg.TmpPath, cfg.Bucket, gcpClient, esClient)
	if newStorageErr != nil {
		logrus.WithError(newStorageErr).Fatal("new storage client")
	}

	speedupAudo := NewSpeedupAudio(contentStorage, cfg.TmpPath, cfg.Atempo)

	scrape := NewScrape(contentStorage, uint32(cfg.MinTextBlockSize), cfg.TmpPath, cfg.TitleLengthLimit)
	text2mp3 := NewText2Mp3(contentStorage, cfg.TmpPath, cfg.WordsPerMinute, cfg.EspeakVoice)
	youtube := NewYoutube(scrape, contentStorage, speedupAudo, cfg.TmpPath)
	contentProcessor := NewRequestProcessor(cfg.TmpPath, scrape, youtube, text2mp3, speedupAudo, cfg.TitleLengthLimit)

	bot := NewBot(cfg.SlackToken, contentProcessor, cfg.LocalPath, cfg.TmpPath, cfg.ChownTo, cfg.Channels)
	bot.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	func() {
		for sig := range c {
			// sig is a ^C, handle it
			logrus.Infof("got signal %d, exiting", sig)
			os.Exit(0)
		}
	}()
}
