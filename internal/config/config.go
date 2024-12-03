package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Name        string `yaml:"name" env-default:"L0"`
	Version     string `yaml:"version" env-default:"1.0.0"`
	StoragePath string `yaml:"storage_path" env-default:"./storage.db"`
	Env         string `yaml:"env" env-default:"local"`
}

// PostgresConfig represents the PostgreSQL configuration
type PostgresConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5433"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	Name     string `yaml:"name" env-default:"l0"`
	PGDriver string `yaml:"pg_driver" env-default:"pq"`
}

// HTTPConfig represents the HTTP server configuration
type HTTPConfig struct {
	Host string `yaml:"host" env-default:"localhost"`
	Port string `yaml:"port" env-default:"8080"`
}

// KafkaConfig represents the Kafka configuration
type KafkaConfig struct {
	BootstrapServers string `yaml:"bootstrap_servers" env-default:"localhost:9092"`
	Topic            string `yaml:"topic" env-default:"orders"`
}

type RedisConfig struct {
	Host string `yaml:"host" env-default:"localhost"`
	Port string `yaml:"port" env-default:"6379"`
}

// Config represents the overall configuration
type Config struct {
	App      AppConfig      `yaml:"app"`
	Postgres PostgresConfig `yaml:"postgres"`
	HTTP     HTTPConfig     `yaml:"http"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	Redis    RedisConfig    `yaml:"redis"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Print("CONFIG_PATH is not set")
		configPath = "config/local.yaml"
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
