package config

import (
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
		VertexAILocation:        os.Getenv("VERTEXAI_LOCATION"),
		VertexAIProject:         os.Getenv("VERTEXAI_PROJECT"),
		VertexAIAvailableModels: customModels,
		VertexAnthropicVersion: os.Getenv("VERTEXAI_ANTHROPIC_VERSION"),
		OpenAIAPIKey:            os.Getenv("OPENAI_API_KEY"),
	}
	return nil
}
