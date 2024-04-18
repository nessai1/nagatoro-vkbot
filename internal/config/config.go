package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Address     string `json:"address"`
	PrepromtDir string `json:"preprompt_dir"`

	VK     VKConfig     `json:"vk"`
	OpenAI OpenAIConfig `json:"openai"`
}

type VKConfig struct {
	GroupAPIKey string `json:"group_api_key"`
}

type OpenAIConfig struct {
	APIToken       string `json:"api_token"`
	ProjectID      string `json:"project_id"`
	OrganizationID string `json:"organization_id"`
}

func ReadConfig(configPath string) (Config, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("cannot open config file: %w", err)
	}
	defer f.Close()

	b := bytes.Buffer{}
	_, err = b.ReadFrom(f)
	if err != nil {
		return Config{}, fmt.Errorf("cannot read config file: %w", err)
	}

	c := Config{}
	err = json.Unmarshal(b.Bytes(), &c)
	if err != nil {
		return Config{}, fmt.Errorf("cannot unmarshal config file: %w", err)
	}

	return c, nil
}
