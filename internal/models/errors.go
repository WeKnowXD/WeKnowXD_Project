package models

type ValidationError struct {
	Loc  []any  `json:"loc"`
	Msg  string `json:"msg"`
	Type string `json:"type"`
}

type HTTPValidationError struct {
	Detail []ValidationError `json:"detail"`
}

type AuthResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}
