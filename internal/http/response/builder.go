package response

import (
	"path/filepath"
	"runtime"
)

func New(code int) *APIResponse {
	return &APIResponse{
		StatusCode: code,
		Status:     StatusFromCode(code),
	}
}

func (r *APIResponse) WithMessage(message string) *APIResponse {
	r.Message = message
	return r
}

func (r *APIResponse) WithData(data any) *APIResponse {
	r.Data = data
	return r
}

func (r *APIResponse) WithErrors(errors any) *APIResponse {
	r.Errors = errors
	return r
}

func (r *APIResponse) WithUserMessage(title, text string) *APIResponse {
	r.UserMessageTitle = title
	r.UserMessageText = text
	return r
}

func (r *APIResponse) WithTag(tag string) *APIResponse {
	r.Tag = tag
	return r
}

// Attach file & line info (use ONLY for 500s)
func (r *APIResponse) WithCaller(skip int) *APIResponse {
	_, file, line, ok := runtime.Caller(skip)
	if ok {
		r.File = filepath.Base(file)
		r.Line = line
	}
	return r
}
