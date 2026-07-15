package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"proxy-gateway/config"
	"proxy-gateway/logger"
	"proxy-gateway/models"
	"proxy-gateway/transformer"

	"golang.org/x/oauth2/google"
	"google.golang.org/genai"
)

func HandleModels(w http.ResponseWriter, r *http.Request) {
	currentTime := time.Now().Unix()
	var allModels []models.Model

	modelIDs := config.AppConfig.VertexAIAvailableModels
	for _, id := range modelIDs {
		allModels = append(allModels, models.Model{
			ID:      id,
			Object:  "model",
			Created: currentTime,
			OwnedBy: "google",
		})
	}

	openaiModels, err := transformer.FetchOpenAIModels()
	if err != nil {
		logger.Logger.Error("HandleModels: Failed to fetch OpenAI models", "error", err)
	} else {
		allModels = append(allModels, openaiModels...)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ModelList{Object: "list", Data: allModels})
}

func HandleGeminiStream(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Debug("HandleGeminiStream: Request received")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var openAIReq map[string]any
	if err := json.NewDecoder(r.Body).Decode(&openAIReq); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	modelName, _ := openAIReq["model"].(string)
	cleanModelName := strings.TrimPrefix(modelName, "google/")

	ctx := r.Context()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  config.AppConfig.VertexAIProject,
		Location: config.AppConfig.VertexAILocation,
	})

	if err != nil {
		logger.Logger.Error("GenAI SDK: Failed to create client", "error", err)
		http.Error(w, "Failed to initialize Vertex AI client", http.StatusInternalServerError)
		return
	}
	var contents []*genai.Content
	if messages, ok := openAIReq["messages"].([]any); ok {
		for _, m := range messages {
			msgMap, ok := m.(map[string]any)
			if !ok {
				continue
			}
			role, _ := msgMap["role"].(string)
			content, _ := msgMap["content"].(string)

			geminiRole := "user"
			if role == "assistant" {
				geminiRole = "model"
			}

			contents = append(contents, &genai.Content{
				Role:  geminiRole,
				Parts: []*genai.Part{{Text: content}},
			})
		}
	}

	flusher, ok := w.(http.Flusher)
	responseStream := client.Models.GenerateContentStream(ctx, cleanModelName, contents, nil)
	for resp, err := range responseStream {
		if err != nil {
			logger.Logger.Error("GenAI SDK: Stream error", "error", err)
			return
		}

		if resp == nil {
			continue
		}

		if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
			for _, part := range resp.Candidates[0].Content.Parts {
				if part.Text != "" {
					openAIChunk := map[string]any{
						"id":      "chatcmpl-gemini",
						"object":  "chat.completion.chunk",
						"created": time.Now().Unix(),
						"model":   modelName,
						"choices": []map[string]any{{
							"index":         0,
							"delta":         map[string]any{"content": part.Text},
							"finish_reason": nil,
						}},
					}

					chunkBytes, _ := json.Marshal(openAIChunk)
					fmt.Fprintf(w, "data: %s\n\n", string(chunkBytes))
					if ok {
						flusher.Flush()
					}
				}
			}
		}
	}

	fmt.Fprint(w, "data: [DONE]\n\n")
	if ok {
		flusher.Flush()
	}
}

func HandleClaudeStream(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Debug("HandleClaudeStream: Request received")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var openAIReq map[string]any
	if err := json.NewDecoder(r.Body).Decode(&openAIReq); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	modelName, _ := openAIReq["model"].(string)
	cleanModelName := strings.TrimPrefix(modelName, "anthropic/")
	cleanModelName = strings.TrimSpace(cleanModelName)

	var systemPrompt string
	var anthropicMessages []map[string]any

	if messages, ok := openAIReq["messages"].([]any); ok {
		for _, m := range messages {
			msg, ok := m.(map[string]any)
			if !ok {
				continue
			}

			role, _ := msg["role"].(string)
			content, _ := msg["content"].(string)

			if role == "system" {
				systemPrompt += content + "\n"
				continue
			}

			anthropicRole := role
			if role != "assistant" && role != "user" {
				anthropicRole = "user"
			}

			anthropicMessages = append(anthropicMessages, map[string]any{
				"role":    anthropicRole,
				"content": content,
			})
		}
	}

	maxTokens := 1024
	if mt, ok := openAIReq["max_tokens"].(float64); ok {
		maxTokens = int(mt)
	}

	anthropicReq := map[string]any{
		"anthropic_version": config.AppConfig.VertexAnthropicVersion,
		"messages":          anthropicMessages,
		"max_tokens":        maxTokens,
		"stream":            true,
	}

	if systemPrompt != "" {
		anthropicReq["system"] = strings.TrimSpace(systemPrompt)
	}

	reqBytes, _ := json.Marshal(anthropicReq)

	location := config.AppConfig.VertexAILocation
	project := config.AppConfig.VertexAIProject

	var endpoint string
	if location == "global" || location == "" {
		endpoint = fmt.Sprintf("https://aiplatform.googleapis.com/v1/projects/%s/locations/global/publishers/anthropic/models/%s:streamRawPredict", project, cleanModelName)
	} else {
		endpoint = fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/anthropic/models/%s:streamRawPredict", location, project, location, cleanModelName)
	}
	logger.Logger.Debug("Claude: Routing stream request", "endpoint", endpoint, "model", cleanModelName)

	req, err := http.NewRequestWithContext(r.Context(), "POST", endpoint, bytes.NewBuffer(reqBytes))
	if err != nil {
		logger.Logger.Error("Claude: Failed to create HTTP request", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	tokenSource, err := google.DefaultTokenSource(r.Context(), "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		logger.Logger.Error("Claude: Failed to create token source", "error", err)
		http.Error(w, "Internal auth configuration error", http.StatusInternalServerError)
		return
	}

	tok, err := tokenSource.Token()
	if err != nil {
		logger.Logger.Error("Claude: Failed to fetch oauth2 access token", "error", err)
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Logger.Error("Claude: Request to Vertex AI failed", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Logger.Error("Claude: API Error", "status", resp.Status, "body", string(body))
		http.Error(w, string(body), resp.StatusCode)
		return
	}

	flusher, ok := w.(http.Flusher)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			dataStr := strings.TrimPrefix(line, "data: ")
			if dataStr == "[DONE]" {
				continue
			}

			var data map[string]any
			if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
				continue
			}

			eventType, _ := data["type"].(string)
			if eventType == "content_block_delta" {
				delta, _ := data["delta"].(map[string]any)
				if deltaType, _ := delta["type"].(string); deltaType == "text_delta" {
					text, _ := delta["text"].(string)

					openAIChunk := map[string]any{
						"id":      "chatcmpl-claude",
						"object":  "chat.completion.chunk",
						"created": time.Now().Unix(),
						"model":   modelName,
						"choices": []map[string]any{{
							"index":         0,
							"delta":         map[string]any{"content": text},
							"finish_reason": nil,
						}},
					}

					chunkBytes, _ := json.Marshal(openAIChunk)
					fmt.Fprintf(w, "data: %s\n\n", string(chunkBytes))
					if ok {
						flusher.Flush()
					}
				}
			}
		}
	}

	fmt.Fprint(w, "data: [DONE]\n\n")
	if ok {
		flusher.Flush()
	}
}
