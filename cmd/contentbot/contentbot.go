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
	SlackToken       string `env:"SLACK_TOKEN"`
	AwsRegion        string `env:"AWS_REGION" envDefault:"us-east-1"`
	MinTextBlockSize int    `env:"MIN_TEXT_BLOCK_SIZE" envDefault:"100"`
	S3Bucket         string `env:"S3_BUCKET"`
	LocalPath        string `env:"LOCAL_PATH" envDefault:"/tmp"`
	WordsPerMinute   int    `env:"WORDS_PER_MINUTE" envDefault:"350"`
	EspeakVoice      string `env:"ESPEAK_VOICE" envDefault:"f5"`
}

var (
	cfg config
)

func main() {
	var err error
	cfg = config{}
	err = env.Parse(&cfg)
	CheckError(err)

	logrus.SetLevel(logrus.DebugLevel)

	s3svc := s3.New(session.New(aws.NewConfig().WithRegion(cfg.AwsRegion).WithCredentials(credentials.NewEnvCredentials())))

	contentStorage := NewContentStorage(s3svc, cfg.LocalPath, cfg.S3Bucket)

	speedupAudo := NewSpeedupAudio(contentStorage, cfg.LocalPath)

	scrape := NewScrape(contentStorage, uint32(cfg.MinTextBlockSize), cfg.LocalPath)
	text2mp3 := NewText2Mp3(contentStorage, cfg.LocalPath, cfg.WordsPerMinute, cfg.EspeakVoice)
	youtube := NewYoutube(scrape, contentStorage, speedupAudo, cfg.LocalPath)
	contentProcessor := NewRequestProcessor(cfg.LocalPath, scrape, youtube, text2mp3, speedupAudo)

	bot := NewBot(cfg.SlackToken, contentProcessor)
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
