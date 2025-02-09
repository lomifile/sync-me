package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type Config struct {
	Host             string `json:"host"`
	Port             string `json:"port"`
	User             string `json:"username"`
	Password         string `json:"password"`
	SyncFilePath     string `json:"sync_file_path"`
	ReceiverFilePath string `json:"receiver_file_path"`
	RootName         string `json:"root_name"`
}

func LoadConfig(src string) *Config {
	var root Config
	bytes, err := os.ReadFile(src)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Println("-- [ERROR]: You don't have any fonfig")
		}
		panic(err)
	}

	json.Unmarshal(bytes, &root)
	return &root
}
