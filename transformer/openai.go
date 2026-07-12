package transformer

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"proxy-gateway/models"
)

func ShouldUseResponsesAPI(model string) bool {
	return strings.HasPrefix(model, "gpt-5")
}
func InjectStreamOptions(model string, raw map[string]any) {
	if _, ok := raw["stream_options"]; !ok {
		if stream, ok := raw["stream"].(bool); stream && ok {
			raw["stream_options"] = map[string]any{
				"include_usage": true,
			}
		}
	}
}

func TransformUsageOpenAI5(raw map[string]any) map[string]any {
	if usage, ok := raw["usage"].(map[string]any); ok {
		return usage
	}
	return map[string]any{
		"prompt_tokens":     0,
		"completion_tokens": 0,
		"total_tokens":      0,
	}
}

func TransformGPT5StreamChunk(raw map[string]any, modelName string) map[string]any {
	eventType, _ := raw["type"].(string)

	if eventType == "response.output_text.delta" {
		deltaText, _ := raw["delta"].(string)

		return map[string]any{
			"id":      "chatcmpl-gpt5",
			"object":  "chat.completion.chunk",
			"created": time.Now().Unix(),
			"model":   modelName,
			"choices": []map[string]any{
				{
					"index": 0,
					"delta": map[string]any{
						"content": deltaText,
					},
					"finish_reason": nil,
				},
			},
		}
	}

	return nil
}

func FetchOpenAIModels() ([]models.Model, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return []models.Model{}, nil
	}

	req, _ := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []models.Model `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}
