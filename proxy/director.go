package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"proxy-gateway/config"
	"proxy-gateway/logger"
	"proxy-gateway/models"
	"proxy-gateway/transformer"
)

func HandleDirector(req *http.Request, target *url.URL) {
	logger.Logger.Debug("Director: Processing request", "method", req.Method, "path", req.URL.Path)

	req.Header.Del("Accept-Encoding")

	state := &ProxyState{}

	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.Host = target.Host

	originalPath := req.URL.Path
	if (originalPath == "/v1/chat/completions" || originalPath == "/v1/embeddings") && req.Body != nil && req.Body != http.NoBody {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			logger.Logger.Error("Director: Error reading request body", "error", err)
			return
		}

		var reqPayload models.CompletionRequest
		if err := json.Unmarshal(bodyBytes, &reqPayload); err == nil {
			state.ModelName = reqPayload.Model

			var raw map[string]any
			if err := json.Unmarshal(bodyBytes, &raw); err == nil {
				transformer.InjectStreamOptions(reqPayload.Model, raw)
				if transformer.ShouldUseResponsesAPI(reqPayload.Model) {
					delete(raw, "stream_options")

					if messages, ok := raw["messages"].([]any); ok {
						var input []map[string]any
						for _, m := range messages {
							msg, ok := m.(map[string]any)
							if !ok {
								continue
							}

							role, _ := msg["role"].(string)
							content, _ := msg["content"].(string)

							contentType := "input_text"
							if role == "assistant" {
								contentType = "output_text"
							}

							input = append(input, map[string]any{
								"role": role,
								"content": []map[string]any{
									{
										"type": contentType,
										"text": content,
									},
								},
							})
						}

						raw["input"] = input
						delete(raw, "messages")
						logger.Logger.Debug("Director: Restructured payload for Responses API", "model", reqPayload.Model)
					}
				}

				bodyBytes, _ = json.Marshal(raw)
			}

			if strings.HasPrefix(reqPayload.Model, "gpt-") || strings.HasPrefix(reqPayload.Model, "text-embedding-") {
				state.IsOpenAI = true
				setupOpenAIRouting(req, originalPath, reqPayload.Model)
			} else {
				setupDefaultVertexRouting(req, originalPath, target.Path)
			}
		}
		req.Header.Del("Accept-Encoding")

		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		req.ContentLength = int64(len(bodyBytes))
		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))
	} else {
		setupDefaultVertexRouting(req, originalPath, target.Path)
	}

	*req = *req.WithContext(context.WithValue(req.Context(), ProxyStateKey, state))
}

func setupOpenAIRouting(req *http.Request, originalPath, model string) {
	req.URL.Scheme = "https"
	req.URL.Host = "api.openai.com"
	req.Host = req.URL.Host

	if transformer.ShouldUseResponsesAPI(model) {
		req.URL.Path = "/v1/responses"
		req.Header.Set("X-Proxy-Model", model)
	} else {
		req.URL.Path = originalPath
	}

	req.Header.Set("Authorization", "Bearer "+config.AppConfig.OpenAIAPIKey)
	logger.Logger.Debug("Director: Routing request to OpenAI API", "path", req.URL.Path)
}

func setupDefaultVertexRouting(req *http.Request, originalPath string, targetPath string) {
	suffixPath := strings.TrimPrefix(originalPath, "/v1")

	req.URL.Path = strings.TrimSuffix(targetPath, "/") + suffixPath
	if tok, err := transformer.GetToken(req.Context()); err == nil {
		req.Header.Set("Authorization", "Bearer "+tok)
	} else {
		logger.Logger.Error("Director: Failed to fetch Vertex AI OAuth token", "error", err)
	}

	logger.Logger.Debug("Director: Routing request to Vertex AI Platform", "path", req.URL.Path)
}
