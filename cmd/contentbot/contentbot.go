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
	. "github.com/kai5263499/rhema/domain"
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
	var err error
	cfg = config{}
	err = env.Parse(&cfg)
	CheckError(err)

	level, err := logrus.ParseLevel(cfg.LogLevel)
	CheckError(err)
	logrus.SetLevel(level)

	esClient, err := elastic.NewClient(elastic.SetURL(cfg.ElasticSearchAddress))
	CheckError(err)

	ctx := context.Background()
	gcpClient, err := storage.NewClient(ctx)
	CheckError(err)

	contentStorage, err := NewContentStorage(cfg.TmpPath, cfg.Bucket, gcpClient, esClient)

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
