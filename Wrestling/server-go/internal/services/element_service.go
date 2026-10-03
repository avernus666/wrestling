package services

import (
	"context"

	"wrestling/internal/models"
	"wrestling/internal/ports"
)

type ElementFilter struct {
	Category    string
	Subcategory string
	Difficulty  string
	Search      string
}

type ElementService struct{ repo ports.ElementRepository }

func NewElementService(repo ports.ElementRepository) *ElementService {
	return &ElementService{repo: repo}
}

func (s *ElementService) List(ctx context.Context, filter ElementFilter) ([]models.Element, error) {
	return s.repo.All(ctx, ports.ElementQuery{
		Category: filter.Category, Subcategory: filter.Subcategory,
		Difficulty: filter.Difficulty, Search: filter.Search,
	})
}

func (s *ElementService) Get(ctx context.Context, id int) (models.Element, bool, error) {
	return s.repo.ByID(ctx, id)
}

func (s *ElementService) ByCategory(ctx context.Context, category string) ([]models.Element, error) {
	return s.repo.ByCategory(ctx, category)
}

func (s *ElementService) Stats(ctx context.Context) (models.Stats, error) {
	return s.repo.Stats(ctx)
}
