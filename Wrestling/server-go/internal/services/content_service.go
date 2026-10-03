package services

import (
	"context"

	"wrestling/internal/models"
	"wrestling/internal/ports"
)

type ContentService struct {
	repo ports.ContentRepository
}

func NewContentService(repo ports.ContentRepository) *ContentService {
	return &ContentService{repo: repo}
}

func (s *ContentService) Bars(ctx context.Context) ([]models.Bar, error) {
	return s.repo.Bars(ctx)
}

func (s *ContentService) Base(ctx context.Context) ([]models.BaseBlock, error) {
	return s.repo.Base(ctx)
}

func (s *ContentService) Safety(ctx context.Context) ([]models.SafetyItem, error) {
	return s.repo.Safety(ctx)
}

func (s *ContentService) Community(ctx context.Context, category, search string, limit, offset int) ([]models.CommunityElement, error) {
	return s.repo.ListCommunity(ctx, category, search, limit, offset)
}

func (s *ContentService) CreateCommunity(ctx context.Context, item models.CommunityElement) error {
	return s.repo.CreateCommunity(ctx, item)
}
func (s *ContentService) Comments(ctx context.Context, elementID string, limit int) ([]models.CommunityComment, error) {
	return s.repo.ListComments(ctx, elementID, limit)
}
func (s *ContentService) CreateComment(ctx context.Context, comment models.CommunityComment) error {
	return s.repo.CreateComment(ctx, comment)
}

func (s *ContentService) Report(ctx context.Context, report models.CommunityReport, reporterID string) error {
	return s.repo.CreateReport(ctx, report, reporterID)
}
