package models

type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type ModelList struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

type CompletionRequest struct {
	Model string `json:"model"`
	Tools []struct {
		GoogleSearch any `json:"google_search,omitempty"`
	} `json:"tools,omitempty"`
}
