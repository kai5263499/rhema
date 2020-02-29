package main

import (
	"os"
	"os/signal"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/caarlos0/env"
	"github.com/sirupsen/logrus"

	. "github.com/kai5263499/rhema"
	. "github.com/kai5263499/rhema/domain"
)

type config struct {
	SlackToken       string   `env:"SLACK_TOKEN"`
	AwsDefaultRegion string   `env:"AWS_DEFAULT_REGION" envDefault:"us-east-1"`
	MinTextBlockSize int      `env:"MIN_TEXT_BLOCK_SIZE" envDefault:"100"`
	S3Bucket         string   `env:"S3_BUCKET"`
	TmpPath          string   `env:"TMP_PATH" envDefault:"/tmp"`
	WordsPerMinute   int      `env:"WORDS_PER_MINUTE" envDefault:"350"`
	EspeakVoice      string   `env:"ESPEAK_VOICE" envDefault:"f5"`
	LocalPath        string   `env:"LOCAL_PATH" envDefault:"/data"`
	Atempo           float32  `env:"ATEMPO" envDefault:"2.0"`
	ChownTo          int      `env:"CHOWN_TO" envDefault:"1000"`
	LogLevel         string   `env:"LOG_LEVEL" envDefault:"info"`
	Channels         []string `env:"CHANNELS" envDefault:"content"`
	TitleLengthLimit int      `env:"TITLE_LENGTH_LIMIT" envDefault:"40"`
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

	s3svc := s3.New(session.New(aws.NewConfig().WithRegion(cfg.AwsDefaultRegion).WithCredentials(credentials.NewEnvCredentials())))

	contentStorage := NewContentStorage(s3svc, cfg.TmpPath, cfg.S3Bucket)

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
