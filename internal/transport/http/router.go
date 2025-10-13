package http

import (
	"net/http"
	"shorted/internal/service/shortener"
	"shorted/internal/transport/http/handlers"
)

type Router struct {
	mux *http.ServeMux
}

func NewRouter(shortenerService *shortener.Service) *Router {
	r := &Router{mux: http.NewServeMux()}
	shortHandler := handlers.NewShortenerHandler(shortenerService)
	r.registerShortenerRoutes(shortHandler)

	return r
}

func (r *Router) registerShortenerRoutes(h *handlers.ShortenerHandler) {
	r.mux.HandleFunc("POST /api/shorten", h.CreateShortURL)
	r.mux.HandleFunc("GET /{code}", h.Redirect)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
