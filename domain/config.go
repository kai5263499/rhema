package domain

type Config struct {
	SubmittedWith       string   `env:"SUBMITTED_WITH" envDefault:"api"`
	ApiHttpPort         int      `env:"API_HTTP_PORT" envDefault:"8080"`
	MinTextBlockSize    uint32   `env:"MIN_TEXT_BLOCK_SIZE" envDefault:"100"`
	WordsPerMinute      uint32   `env:"WORDS_PER_MINUTE" envDefault:"350"`
	EspeakVoice         string   `env:"ESPEAK_VOICE" envDefault:"f5"`
	Atempo              string   `env:"ATEMPO" envDefault:"2.0"`
	TitleLengthLimit    uint32   `env:"TITLE_LENGTH_LIMIT" envDefault:"40"`
	ChownTo             int      `env:"CHOWN_TO" envDefault:"1000"`
	LogLevel            string   `env:"LOG_LEVEL" envDefault:"debug"`
	LocalPath           string   `env:"LOCAL_PATH" envDefault:"./.docker/data"`
	TmpPath             string   `env:"TMP_PATH" envDefault:"/tmp"`
	RedisHost           string   `env:"REDIS_HOST" envDefault:"localhost"`
	RedisPort           string   `env:"REDIS_PORT" envDefault:"6379"`
	KafkaBrokers        string   `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
	KafkaRequestTopic   string   `env:"KAFKA_REQUESTS_TOPIC" envDefault:"rhema.requests"`
	KafkaGroupId        string   `env:"KAFKA_GROUP_ID" envDefault:"rhema-processor"`
	BoltDBPath          string   `env:"BOLTDB_PATH"`
	BoltDBBucket        string   `env:"BOLTDB_BUCKET" envDefault:"content"`
	SlackToken          string   `env:"SLACK_TOKEN"`
	Channels            []string `env:"CHANNELS" envDefault:"content"`
	CayleyStoragePath   string   `env:"CAYLEY_STORAGE_PATH"`
	RequestProcessorUri string   `env:"REQUEST_PROCESSOR_URI" envDefault:"http://localhost:8080"`
	DDAgentHost         string   `env:"DD_AGENT_HOST" envDefault:"localhost"`
	DDAgentPort         int      `env:"DD_AGENT_PORT" envDefault:"8125"`
	DiscordBotToken     string   `env:"DISCORD_BOT_TOKEN"`
}
