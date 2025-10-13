package repositories

import "shorted/internal/domain/models"

type LinkRepository interface {
	Save(link *models.Link) error
	FindByCode(shortCode string) (*models.Link, error)
}
