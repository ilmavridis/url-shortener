package config

import (
	"ilmavridis/url-shortener/logger"

	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv" //to load .env
	"github.com/spf13/viper"
)

type server struct {
	Address      string        `mapstructure:"address"`
	TimeoutWrite time.Duration `mapstructure:"timeoutWrite"`
	TimeoutRead  time.Duration `mapstructure:"timeoutRead"`
	TimeoutIdle  time.Duration `mapstructure:"timeoutIdle"`
}

type redis struct {
	Address  string        `mapstructure:"address"`
	Pass     string        `mapstructure:"pass"`
	Database int           `mapstructure:"database"`
	Expiry   time.Duration `mapstructure:"expiry"`
}

// Config holds all service configs
type Config struct {
	Server server
	Redis  redis
}

var configs Config

func Read() error {
	err := godotenv.Load("../" + `.env`)
	if err != nil {
		logger.Error("Could not read .env file: ", err)
	}

	var configFile string

	// Use the test configuration file If it is running in test
	if flag.Lookup("test.v") == nil {
		configFile = os.Getenv("CONFIG_FILE")
	} else {
		configFile = os.Getenv("CONFIG_FILE_TEST")
	}

	v := viper.New()

	// Set configuration file type and directory
	v.SetConfigType("yaml")
	v.AddConfigPath(filepath.Join("../"))
	v.AddConfigPath(filepath.Join("/app")) //Docker

	// If the CONFIG_FILE environment variable is empty, use the default configuration file
	if configFile == "" {
		v.SetConfigName("config-default")
	}

	// Load the configuration values from the user-defined configuration file
	v.SetConfigName(configFile)
	err = v.ReadInConfig()
	if err != nil {
		return err
	}

	v.Unmarshal(&configs)
	return nil

}

func Get() Config {
	return configs
}
