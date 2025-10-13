package contract

import "net/http"

type ResponseWriter interface {
	Write(rw http.ResponseWriter, status int, data interface{})
}
