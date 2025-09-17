package httputil

import (
	"net/http"

	"github.com/bytedance/sonic"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string, details error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := ErrorResponse{
		Code:    statusCode,
		Message: message,
	}

	if details != nil {
		resp.Details = details.Error()
	}

	sonic.ConfigFastest.NewEncoder(w).Encode(resp)
}

func WriteJSONResponse(w http.ResponseWriter, statusCode int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if body != nil {
		sonic.ConfigDefault.NewEncoder(w).Encode(body)
	}
}
