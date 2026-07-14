package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	
	"proxy-gateway/logger"
)

type Config struct {
	Port                    string
	VertexAILocation        string
	VertexAIProject         string
	VertexAnthropicVersion  string
	VertexAIAvailableModels []string
	OpenAIAPIKey            string
}

var AppConfig *Config

func LoadConfig() error {
	if err := godotenv.Load(); err != nil {
		logger.Logger.Debug("Warning: .env file not found, reading from system env instead")
	}

	location := os.Getenv("VERTEXAI_LOCATION")
	projectID := os.Getenv("VERTEXAI_PROJECT")
	if location == "" || projectID == "" {
		return fmt.Errorf("VERTEXAI_LOCATION and VERTEXAI_PROJECT env vars must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	var customModels []string
	if modelsStr := os.Getenv("VERTEXAI_AVAILABLE_MODELS"); modelsStr != "" {
		for _, id := range strings.Split(modelsStr, ",") {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				customModels = append(customModels, trimmed)
			}
		}
	}

	AppConfig = &Config{
		Port:                    port,
		VertexAILocation:        location,
		VertexAIProject:         projectID,
		VertexAIAvailableModels: customModels,
		VertexAnthropicVersion: os.Getenv("VERTEXAI_ANTHROPIC_VERSION"),
		OpenAIAPIKey:            os.Getenv("OPENAI_API_KEY"),
	}
	return nil
}
