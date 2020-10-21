package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/allisonmorgan/tfidf"
	"github.com/caarlos0/env"
	"github.com/davecgh/go-spew/spew"
	"github.com/gomodule/redigo/redis"
	rg "github.com/redislabs/redisgraph-go"
	"github.com/sirupsen/logrus"
)

type config struct {
	LogLevel      string `env:"LOG_LEVEL" envDefault:"info"`
	RedisHost     string `env:"REDIS_HOST"`
	RedisPort     string `env:"REDIS_PORT" envDefault:"6379"`
	RedisGraphKey string `env:"REDIS_GRAPH_KEY" envDefault:"rhema-content"`
}

var (
	cfg config
)

func ReadCSV(filepath string) ([][]string, error) {
	csvfile, err := os.Open(filepath)

	if err != nil {
		fmt.Printf("Unable to read csv: %v", err)
		return nil, err
	}
	defer csvfile.Close()

	reader := csv.NewReader(csvfile)
	fields, err := reader.ReadAll()

	return fields, err
}

func main() {

	cfg = config{}
	if err := env.Parse(&cfg); err != nil {
		logrus.WithError(err).Fatal("parse config")
	}

	if level, err := logrus.ParseLevel(cfg.LogLevel); err != nil {
		logrus.WithError(err).Fatal("parse log level")
	} else {
		logrus.SetLevel(level)
	}

	redisConnStr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	logrus.Debugf("connecting to redis %s", redisConnStr)
	redisConn, redisConnErr := redis.Dial("tcp", redisConnStr)
	if redisConnErr != nil {
		logrus.WithError(redisConnErr).Fatal("unable to connect to redis")
	}

	redisGraph := rg.GraphNew(cfg.RedisGraphKey, redisConn)

	query := fmt.Sprintf("MATCH (a:content) RETURN a.text")
	result, queryErr := redisGraph.Query(query)
	if queryErr != nil {
		logrus.WithError(queryErr).Fatalf("query error")
	}

	frequency := tfidf.NewTermFrequencyStruct()

	cnt := 0
	for result.Next() {
		r := result.Record()

		var text string

		txtI, foundT := r.Get("a.text")
		if foundT {
			text = txtI.(string)
		}

		cnt++
		frequency.AddDocument(text)
	}

	logrus.Infof("processing inverse document frequency from %d documents", cnt)
	frequency.InverseDocumentFrequency()
	logrus.Infof("processed %d terms", len(frequency.InverseDocMap))
	spew.Dump(frequency.InverseDocMap)
}
