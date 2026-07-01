package response

type Response struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
	Error   any  `json:"error,omitempty"`
	Meta    any  `json:"meta,omitempty"`
}
