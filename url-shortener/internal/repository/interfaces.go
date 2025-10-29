package repository

import (
	"context"
	"url-shortener/internal/models"
)

type URLRepository interface {
	Create(ctx context.Context, url *models.URL) error
	GetByID(ctx context.Context, ID int) (*models.URL, error)
	FindByShortCode(ctx context.Context, shortCode string) (*models.URL, error)
	FindByOriginalURL(ctx context.Context, originalURL string) (*models.URL, error)
	Update(ctx context.Context, url *models.URL) error
	Delete(ctx context.Context, ID int) error
}

type AnalyticsRepository interface {
	SaveClick(ctx context.Context, click *models.Click) error
	GetAnalyticsByID(ctx context.Context, ID int) (*models.Analytics, error)
	GetAnalyticsByShortCode(ctx context.Context, shortCode string) (*models.Analytics, error)
}

type CacheRepository interface {
	GetURL(ctx context.Context, shortCode string) (*models.URL, error)
	SetURL(ctx context.Context, url *models.URL) error
	DeleteURL(ctx context.Context, shortCode string) error
}
