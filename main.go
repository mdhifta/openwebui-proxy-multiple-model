package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"proxy-gateway/config"
	"proxy-gateway/logger"
	"proxy-gateway/proxy"
)

func SmartChatRouter(proxyEngine http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/chat/completions" && r.Method == "POST" {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				var payload map[string]any
				if err := json.Unmarshal(bodyBytes, &payload); err == nil {
					modelName, _ := payload["model"].(string)
					if strings.HasPrefix(modelName, "google/gemini") || strings.HasPrefix(modelName, "gemini-") {
						proxy.HandleGeminiStream(w, r)
						return
					}

					if strings.HasPrefix(modelName, "claude-") || strings.HasPrefix(modelName, "anthropic/claude-") {
						proxy.HandleClaudeStream(w, r)
						return
					}
				}
			}
		}

		proxyEngine.ServeHTTP(w, r)
	}
}

func main() {
	logger.InitSlogLogger()

	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Initialization failed: %v", err)
	}

	targetURL := config.AppConfig.VertexAILocation
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Invalid upstream URL: %v", err)
	}

	proxyEngine := proxy.NewMakeProxy(target)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models", proxy.HandleModels)
	mux.Handle("/", SmartChatRouter(proxyEngine))

	addr := ":" + config.AppConfig.Port
	logger.Logger.Info("Proxy Engine online", "listen_port", config.AppConfig.Port, "upstream_target", targetURL)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server shutdown: %v", err)
	}
}
