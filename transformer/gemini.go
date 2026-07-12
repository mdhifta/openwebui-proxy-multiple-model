package transformer

import (
	"context"
	"encoding/json"
)

func OpenAIToGemini3(openAIBody []byte) ([]byte, error) {
	var raw map[string]any
	if err := json.Unmarshal(openAIBody, &raw); err != nil {
		return openAIBody, err
	}

	geminiPayload := map[string]any{
		"contents": raw["messages"],
	}

	return json.Marshal(geminiPayload)
}

func TransformUsageGemini3(raw map[string]any) map[string]any {
	if meta, ok := raw["usageMetadata"].(map[string]any); ok {
		pTokens, _ := meta["promptTokenCount"].(float64)
		cTokens, _ := meta["candidatesTokenCount"].(float64)
		tTokens, _ := meta["totalTokenCount"].(float64)

		return map[string]any{
			"prompt_tokens":     int(pTokens),
			"completion_tokens": int(cTokens),
			"total_tokens":      int(tTokens),
		}
	}
	return map[string]any{"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}
}

func GetToken(ctx context.Context) (string, error) {
	return "MOCK_VERTEX_OAUTH_TOKEN", nil
}
