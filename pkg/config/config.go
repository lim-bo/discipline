package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var (
	once     sync.Once
	instance *Config
)

type Config struct {
}

func New() *Config {
	once.Do(func() {
		err := godotenv.Load("./configs/.env")
		if err != nil {
			log.Fatal("loading envs error: ", err)
		}
		instance = &Config{}
	})
	return instance
}

func (c *Config) GetString(key string) string {
	return os.Getenv(key)
}
