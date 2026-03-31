package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel     string           `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`
	WordsAddress string           `yaml:"words_address" env:"WORDS_ADDRESS"`
	HTTPServer   HTTPServerConfig `yaml:"http_server"`
}

type HTTPServerConfig struct {
	Address string        `yaml:"address" env:"HTTP_SERVER_ADDRESS" env-default:":8080"`
	Timeout time.Duration `yaml:"timeout" env:"HTTP_SERVER_TIMEOUT" env-default:"5s"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	var err error
	if configPath != "" {
		err = cleanenv.ReadConfig(configPath, &cfg)
	} else {
		err = cleanenv.ReadEnv(&cfg)
	}
	if err != nil {
		log.Fatalf("cannot read config: %s", err)
	}
	return cfg
}
