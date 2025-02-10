package repository

import (
	"api/config"
)

type Repository struct{ cfg *config.Config }
type RepositoryOption func(*Repository)

func WithConfig(cfg *config.Config) RepositoryOption {
	return func(r *Repository) {
		r.cfg = cfg
	}
}
func NewRepository(opts ...RepositoryOption) *Repository {
	r := &Repository{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}
