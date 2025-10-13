package memory

import (
	"shorted/internal/domain/models"
	"shorted/internal/domain/repositories"
	"sync"
)

type LinkRepo struct {
	mu    sync.Mutex
	links map[string]*models.Link
}

func NewLinkRepo() *LinkRepo {
	return &LinkRepo{
		links: make(map[string]*models.Link),
	}
}

func (r *LinkRepo) Save(link *models.Link) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.links[link.ShortCode] = link
	return nil
}

func (r *LinkRepo) FindByCode(shortCode string) (*models.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	link, exists := r.links[shortCode]
	if !exists {
		return nil, repositories.ErrNotFound

	}

	return link, nil
}
