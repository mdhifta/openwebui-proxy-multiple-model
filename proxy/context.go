package proxy

type contextKey string

const ProxyStateKey contextKey = "proxy_state"

type ProxyState struct {
	IsOpenAI  bool
	IsGemini3 bool
	ModelName string
}
