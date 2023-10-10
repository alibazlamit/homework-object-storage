package config

import (
	"os"

	"github.com/alibazlamit/homework-object-storage/logger"
)

type Config struct {
	MainBucket  string
	StoragePort string
}

var AppConfig Config

func init() {
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

}
