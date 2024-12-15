package config

import (
	"os"
	_ "github.com/joho/godotenv/autoload"
	"strconv"
)

type AuthConfig struct {
	SessionKey 			string
	TimeoutInSeconds 	int
}

type Config struct {
	MainServerEndpoint string
	AuthConfig AuthConfig
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		MainServerEndpoint: os.Getenv("MAIN_SERVER_ENDPOINT"),
		AuthConfig: AuthConfig{
			SessionKey: os.Getenv("SESSION_KEY"),
			TimeoutInSeconds: func() int {
				value, _ := strconv.Atoi(os.Getenv("TIMEOUT_IN_SECONDS"))
				return value
			}(),
		},
	}

	return cfg, nil
}
