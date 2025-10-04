package config

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	HttpAddress  string        `yaml:"httpAddress" env-default:":80"`
	HttpsAddress string        `yaml:"httpsAddress" env-default:":443"`
	Timeout      time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

var (
	instance *Config
	once     sync.Once
)

func MustLoad() *Config {
	once.Do(func() {
		currentDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("path error: %s", err)
		}
		configPath := filepath.Join(currentDir, "config", "local.yaml")
		configPath = filepath.Clean(configPath)

		if configPath == "" {
			log.Fatal("configPath is empty")
		}

		if _, err := os.Stat(configPath); err != nil {
			log.Fatalf("error opening config file: %s", err)
		}

		var cfg Config
		err = cleanenv.ReadConfig(configPath, &cfg)
		if err != nil {
			log.Fatalf("error reading config file: %s", err)
		}

		instance = &cfg
	})

	return instance
}

// return global config instance
func Get() *Config {
	if instance == nil {
		log.Fatal("config not initialized. call MustLoad()")
	}
	return instance
}
