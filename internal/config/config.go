package config

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AppEnv       string `yaml:"env" env-default:"local"`
	BotPrefix    string `yaml:"bot_preifx" env-default:"!"`
	DiscordToken string `yaml:"discord_token" env-required:"true"`
	YoutubeKey   string `yaml:"youtube_key" env-required:"true"`
	SentryDSN    string `yaml:"sentry_dsn" env-required:"true"`
}

var Dg *discordgo.Session

const (
	ENV_LOCAL      = "LOCAL"
	ENV_TEST       = "TEST"
	ENV_PRODUCTION = "PROD"
)

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		path, err := os.Getwd()
		if err != nil {
			log.Fatalf("error getting current directory path: %s", err)
		}

		configPath = path + "config/local.yaml"
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}

func Initialize(discordToken string) {
	var err error

	Dg, err = discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatalln("ERROR: error creating Discord session ", err)
	}
}

func OpenConnection() {
	if err := Dg.Open(); err != nil {
		log.Fatalln("ERROR: unable to open connection,", err)
	}
}
