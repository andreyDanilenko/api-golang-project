package apierror

import (
	"encoding/json"
	"log"
	"net/http"
	"shorted/internal/contract"
)

type writer struct{}

func New() contract.ErrorWriter {
	return &writer{}
}

func (w *writer) WriteError(rw http.ResponseWriter, status int, message string) {
	w.WriteWithCode(rw, status, http.StatusText(status), message, nil)
}

func (w *writer) WriteWithCode(rw http.ResponseWriter, status int, errorCode, message string, details interface{}) {
	if status >= 500 {
		log.Printf("internal error: %s", message)
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)

	json.NewEncoder(rw).Encode(struct {
		Code    int         `json:"code"`
		Error   string      `json:"error"`
		Message string      `json:"message"`
		Details interface{} `json:"details,omitempty"`
	}{
		Code:    status,
		Error:   errorCode,
		Message: message,
		Details: details,
	})

}
