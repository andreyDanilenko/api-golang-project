package shortener

import "shorted/internal/domain/repositories"

type Service struct {
	repo repositories.LinkRepository
}

func NewService(repo repositories.LinkRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateShortURL() {
	// структура данных
	// рера
	// ответ
}

func (s *Service) GetOriginalURL() {
	//репа
	//ответ
}
