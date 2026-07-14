package proxy

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"proxy-gateway/transformer"
	"proxy-gateway/logger"
)

func HandleModifyResponse(resp *http.Response) error {
	logger.Logger.Debug("Response! Status: %d, Path: %s\n", resp.StatusCode, resp.Request.URL.Path)

	state, ok := resp.Request.Context().Value(ProxyStateKey).(*ProxyState)
	if !ok {
		logger.Logger.Debug("Warning: ProxyState not find in the context\n")
		state = &ProxyState{}
	}

	ct := resp.Header.Get("Content-Type")
	logger.Logger.Debug("Content-Type from OpenAI: %s\n", ct)

	if state.IsGemini3 && strings.Contains(ct, "application/json") {
		return processGemini3Stream(resp)
	}

	if state.IsOpenAI && strings.Contains(ct, "text/event-stream") && strings.Contains(resp.Request.URL.Path, "/v1/responses") {
		logger.Logger.Debug("Condition GPT-5 success, Model: %s\n", state.ModelName)

		if !strings.Contains(ct, "text/event-stream") {
			logger.Logger.Debug("Error from OpenAI! OpenAI did not return a stream. Status: %d\n", resp.StatusCode)
		}

		return processOpenAIResponsesStream(resp, state.ModelName)
	}

	logger.Logger.Debug("Skipping ModifyResponse (Normal bypass)\n")
	return nil
}

func processGemini3Stream(resp *http.Response) error {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for decoder.More() {
			var raw map[string]interface{}
			if err := decoder.Decode(&raw); err != nil {
				return
			}

			mappedUsage := transformer.TransformUsageGemini3(raw)
			if mappedUsage != nil {
				raw["usage"] = mappedUsage
			}

			modifiedJSON, _ := json.Marshal(raw)
			_, _ = pw.Write([]byte("data: " + string(modifiedJSON) + "\n\n"))
		}
		_, _ = pw.Write([]byte("data: [DONE]\n\n"))
	}()

	resp.Body = pr
	resp.Header.Set("Content-Type", "text/event-stream")
	return nil
}

func processOpenAIResponsesStream(resp *http.Response, modelName string) error {
	logger.Logger.Debug("Start process OpenAIResponses Stream\n")

	originalBody := resp.Body
	pr, pw := io.Pipe()
	go func() {
		logger.Logger.Debug("Goroutine OpenAI running...\n")
		defer func() {
			logger.Logger.Debug("Goroutine done and closed\n")
			pw.Close()
			originalBody.Close()
		}()

		reader := bufio.NewReader(originalBody)
		for {
			lineBytes, err := reader.ReadBytes('\n')

			if len(lineBytes) > 0 {
				line := string(lineBytes)
				line = strings.TrimRight(line, "\r\n")

				logger.Logger.Debug("[Stream Input] %s\n", line)

				if strings.HasPrefix(line, "data: ") && !strings.Contains(line, "[DONE]") {
					jsonStr := strings.TrimPrefix(line, "data: ")
					var raw map[string]interface{}

					if errUnmarshal := json.Unmarshal([]byte(jsonStr), &raw); errUnmarshal == nil {
						standardChunk := transformer.TransformGPT5StreamChunk(raw, modelName)

						if standardChunk != nil {
							mappedUsage := transformer.TransformUsageOpenAI5(raw)
							if mappedUsage != nil {
								standardChunk["usage"] = mappedUsage
							}

							modifiedJSON, _ := json.Marshal(standardChunk)
							_, _ = pw.Write([]byte("data: " + string(modifiedJSON) + "\n\n"))
						}
						continue
					}
				}

				_, _ = pw.Write([]byte(line + "\n\n"))
			}

			if err != nil {
				if err != io.EOF {
					logger.Logger.Debug("Stream disconnected or error: %v\n", err)
				}
				break
			}
		}
	}()

	resp.Body = pr
	resp.Header.Set("Content-Type", "text/event-stream")
	resp.Header.Del("Content-Encoding")
	
	return nil
}
