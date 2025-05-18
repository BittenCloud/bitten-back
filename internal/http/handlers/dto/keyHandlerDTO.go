package dto

// VlessKeyResponse defines the structure of the JSON response for a VLESS key.
type VlessKeyResponse struct {
	VlessKey string `json:"vless_key"`         // The generated VLESS key string.
	UserID   string `json:"user_id"`           // The ID of the user for whom the key was generated.
	Remarks  string `json:"remarks,omitempty"` // Optional remarks or a name for the key.
}
