package config

import (
	"os"
	"strconv"

	"github.com/alibazlamit/homework-object-storage/logger"
)

type Config struct {
	MainBucket     string
	StoragePort    string
	GeneralTimeout int
}

var AppConfig Config

func (cnf Config) LoadConfig() {
	mainBucket := os.Getenv("MAIN_BUCKET")

	if mainBucket == "" {
		logger.Log.Warn("MAIN_BUCKET environment variable is not set or empty.")
	} else {
		logger.Log.Infof("MAIN_BUCKET: %s\n", mainBucket)
	}
	AppConfig.MainBucket = mainBucket

	port := os.Getenv("STORAGE_PORT")
	if port == "" {
		logger.Log.Warn("STORAGE_PORT environment variable is not set or empty.")
	} else {
		logger.Log.Infof("STORAGE_PORT: %s\n", port)
	}

	AppConfig.StoragePort = port

	timeout := os.Getenv("GENERAL_TIMEOUT")
	if timeout == "" {
		logger.Log.Warn("GENERAL_TIMEOUT environment variable is not set or empty.")
	}
	timeoutVal, err := strconv.Atoi(timeout)
	if err != nil {
		// Handle the error if the conversion fails
		logger.Log.Errorf("Failed to convert GENERAL_TIMEOUT to int: %v", err)
	} else {
		logger.Log.Infof("GENERAL_TIMEOUT: %s\n", timeout)
		AppConfig.GeneralTimeout = timeoutVal
	}

}

func (cnf Config) SetMockEnv(config Config) {
	AppConfig = Config{
		MainBucket:     config.MainBucket,
		StoragePort:    config.StoragePort,
		GeneralTimeout: config.GeneralTimeout,
	}
}
