package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

const configFilename = ".gatorconfig.json"

type Config struct {
	DBUrl string `json:"db_url"`
	CurrentUserName *string `json:"current_user_name"`
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, configFilename), nil
}

func writeConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, jsonData, 0666); err != nil {
		return err
	}

	return nil
}

func Read() (*Config, error) {
	config := &Config{}

	configPath, err := getConfigPath()
	if err != nil {
		return config, err
	}
	
	configFile, err := os.Open(configPath)
	if err != nil {
		return config, err
	}
	defer configFile.Close()

	byteValue, err := io.ReadAll(configFile)
	if err != nil {
		return config, err
	}

	if err := json.Unmarshal(byteValue, &config); err != nil {
		return config, err
	}
	
	return config, nil
}

func (c *Config) SetUser(currentUserName string) error {
	c.CurrentUserName = &currentUserName

	return writeConfig(c)
}
