package node

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	Name           string `json:"name"`
	ContainerdPath string `json:"containerd_path"`
	APIHost        string `json:"api_host"`
	APIPort        int    `json:"api_port"`
}

func LoadConfig(path string) (cfg *Config, err error) {
	file, openErr := os.Open(path)

	if openErr != nil {
		return nil, errors.New("LoadConfig: os.Open failed with error: " + openErr.Error())
	}

	var buff Config
	decodeErr := json.NewDecoder(file).Decode(&buff)

	if decodeErr != nil {
		return nil, errors.New("LoadConfig: Decoder.Decode failed with error: " + decodeErr.Error())
	}

	return &buff, nil
}
