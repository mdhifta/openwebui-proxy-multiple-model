package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                    string
	VertexAILocation        string
	VertexAIProject         string
	VertexAIAvailableModels []string
	OpenAIAPIKey            string
	GeminiAPIKey            string
}

var AppConfig *Config

func LoadConfig() error {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, reading from system env instead")
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
		OpenAIAPIKey:            os.Getenv("OPENAI_API_KEY"),
		GeminiAPIKey:            os.Getenv("GEMINI_API_KEY"),
	}
	return nil
}
