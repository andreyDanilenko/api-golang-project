package apiresponse

import (
	"encoding/json"
	"net/http"
	"shorted/internal/contract"
)

type writer struct{}

func New() contract.ResponseWriter {
	return &writer{}
}

func (w *writer) Write(rw http.ResponseWriter, status int, data interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	json.NewEncoder(rw).Encode(data)
}
