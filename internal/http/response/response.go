package response

type APIResponse struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`

	Message string `json:"message,omitempty"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`

	UserMessageTitle string `json:"userMessageTitle,omitempty"`
	UserMessageText  string `json:"userMessageText,omitempty"`

	Data   any `json:"data,omitempty"`
	Errors any `json:"errors,omitempty"`

	Tag string `json:"tag,omitempty"`
}
