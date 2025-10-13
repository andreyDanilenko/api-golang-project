package contract

import "net/http"

type ErrorWriter interface {
	WriteError(rw http.ResponseWriter, status int, message string)
	WriteWithCode(rw http.ResponseWriter, status int, errorCode, message string, details interface{})
}
