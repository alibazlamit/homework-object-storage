package config

import (
	"fmt"
	"os"
)

type Config struct {
	MainBucket  string
	StoragePort string
}

var AppConfig Config

func init() {
	mainBucket := os.Getenv("MAIN_BUCKET")

	if mainBucket == "" {
		fmt.Println("MAIN_BUCKET environment variable is not set or empty.")
	} else {
		fmt.Printf("MAIN_BUCKET: %s\n", mainBucket)
	}
	AppConfig.MainBucket = mainBucket

	port := os.Getenv("STORAGE_PORT")
	if port == "" {
		fmt.Println("STORAGE_PORT environment variable is not set or empty.")
	} else {
		fmt.Printf("STORAGE_PORT: %s\n", port)
	}

	AppConfig.StoragePort = port

}
