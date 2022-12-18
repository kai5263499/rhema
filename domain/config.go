package domain

type Config struct {
	SubmittedWith     string   `env:"SUBMITTED_WITH" envDefault:"api"`
	ApiHttpPort       int      `env:"API_HTTP_PORT" envDefault:"8080"`
	MinTextBlockSize  uint32   `env:"MIN_TEXT_BLOCK_SIZE" envDefault:"100"`
	WordsPerMinute    uint32   `env:"WORDS_PER_MINUTE" envDefault:"350"`
	EspeakVoice       string   `env:"ESPEAK_VOICE" envDefault:"f5"`
	Atempo            string   `env:"ATEMPO" envDefault:"2.0"`
	TitleLengthLimit  uint32   `env:"TITLE_LENGTH_LIMIT" envDefault:"40"`
	ChownTo           int      `env:"CHOWN_TO" envDefault:"1000"`
	LogLevel          string   `env:"LOG_LEVEL" envDefault:"debug"`
	LocalPath         string   `env:"LOCAL_PATH"`
	RedisHost         string   `env:"REDIS_HOST"`
	RedisPort         string   `env:"REDIS_PORT" envDefault:"6379"`
	KafkaBrokers      string   `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
	KafkaRequestTopic string   `env:"KAFKA_REQUESTS_TOPIC" envDefault:"rhema.requests"`
	KafkaGroupId      string   `env:"KAFKA_GROUP_ID" envDefault:"rhema-processor"`
	BoltDBPath        string   `env:"BOLTDB_PATH"`
	SlackToken        string   `env:"SLACK_TOKEN"`
	Channels          []string `env:"CHANNELS" envDefault:"content"`
}
