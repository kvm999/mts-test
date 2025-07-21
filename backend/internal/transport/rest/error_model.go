package rest

// ErrorResponse represents an error in API responses
// @Description Error response format
type ErrorResponse struct {
	// Error message
	// @Description Human-readable error message
	// @Example "Validation failed"
	Message string `json:"message" example:"Validation failed"`

	// Error code (optional)
	// @Description Machine-readable error code
	// @Example "INVALID_INPUT"
	Code string `json:"code,omitempty" example:"INVALID_INPUT"`
} // @name ErrorResponse
