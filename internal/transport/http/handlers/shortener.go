package handlers

import (
	"net/http"
	"shorted/internal/service/shortener"
)

type ShortenerHandler struct {
	service *shortener.Service
}

func NewShortenerHandler(service *shortener.Service) *ShortenerHandler {
	return &ShortenerHandler{service: service}
}

func (h *ShortenerHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	// обработка тела
	// сервис
	// респотсе
}

func (h *ShortenerHandler) Redirect(w http.ResponseWriter, r *http.Request) {

}
