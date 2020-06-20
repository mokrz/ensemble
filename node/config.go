package node

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds administrative settings
type Config struct {
	Name           string `json:"name"`
	ContainerdPath string `json:"containerd_path"`
	APIHost        string `json:"api_host"`
	APIPort        int    `json:"api_port"`
}

// LoadConfig reads the given .json file into a node.Config instance
func LoadConfig(path string) (cfg *Config, err error) {
	file, openErr := os.Open(path)

	if openErr != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, openErr)
	}

	var buff Config
	decodeErr := json.NewDecoder(file).Decode(&buff)

	if decodeErr != nil {
		return nil, fmt.Errorf("failed to decode %s: %w", path, decodeErr)
	}

	return &buff, nil
}
