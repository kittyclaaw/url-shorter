package service

import (
	"context"
	"url-shortener/internal/models"
	"url-shortener/internal/repository"
)

type AnalyticsService struct {
	clickRepo repository.AnalyticsRepository
	urlRepo   repository.URLRepository
}

func NewAnalyticsService(clickRepo repository.AnalyticsRepository, urlRepo repository.URLRepository) *AnalyticsService {
	return &AnalyticsService{
		clickRepo: clickRepo,
		urlRepo:   urlRepo,
	}
}

func (s *AnalyticsService) SaveClick(ctx context.Context, click *models.Click) error {
	return s.clickRepo.SaveClick(ctx, click)
}

func (s *AnalyticsService) GetAnalyticsByID(ctx context.Context, urlID int) (*models.Analytics, error) {
	return s.clickRepo.GetAnalyticsByID(ctx, urlID)
}

func (s *AnalyticsService) GetAnalyticsByShortCode(ctx context.Context, shortCode string) (*models.Analytics, error) {
	return s.clickRepo.GetAnalyticsByShortCode(ctx, shortCode)
}
