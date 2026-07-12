package proxy

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"proxy-gateway/transformer"
)

func HandleModifyResponse(resp *http.Response) error {
	log.Printf("[DEBUG Modifier] Response! Status: %d, Path: %s\n", resp.StatusCode, resp.Request.URL.Path)

	state, ok := resp.Request.Context().Value(ProxyStateKey).(*ProxyState)
	if !ok {
		log.Printf("[DEBUG Modifier] Warning: ProxyState not find in the context\n")
		state = &ProxyState{}
	}

	ct := resp.Header.Get("Content-Type")
	log.Printf("[DEBUG Modifier] Content-Type from OpenAI: %s\n", ct)

	if state.IsGemini3 && strings.Contains(ct, "application/json") {
		return processGemini3Stream(resp)
	}

	if state.IsOpenAI && strings.Contains(ct, "text/event-stream") && strings.Contains(resp.Request.URL.Path, "/v1/responses") {
		log.Printf("[DEBUG Modifier] Condition GPT-5 success, Model: %s\n", state.ModelName)

		if !strings.Contains(ct, "text/event-stream") {
			log.Printf("[DEBUG Modifier] ERROR DARI OPENAI! OpenAI tidak mengirim stream. Status: %d\n", resp.StatusCode)
		}

		return processOpenAIResponsesStream(resp, state.ModelName)
	}

	log.Printf("[DEBUG Modifier] Melewati ModifyResponse (Bypass normal)\n")
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
	log.Printf("[DEBUG Stream] Start processOpenAIResponsesStream\n")

	originalBody := resp.Body
	pr, pw := io.Pipe()
	go func() {
		log.Printf("[DEBUG Goroutine] Goroutine OpenAI running...\n")
		defer func() {
			log.Printf("[DEBUG Goroutine] Goroutine done and closed\n")
			pw.Close()
			originalBody.Close()
		}()

		reader := bufio.NewReader(originalBody)
		for {
			lineBytes, err := reader.ReadBytes('\n')

			if len(lineBytes) > 0 {
				line := string(lineBytes)
				line = strings.TrimRight(line, "\r\n")

				log.Printf("[DEBUG Stream Input] %s\n", line)

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
					log.Printf("[ERROR Reader] Stream disconnected or error: %v\n", err)
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
