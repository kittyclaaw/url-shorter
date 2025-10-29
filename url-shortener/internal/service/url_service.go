package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/redis/go-redis/v9"
	"log"
	"math/rand"
	"net/url"
	"strings"
	"url-shortener/internal/models"
	"url-shortener/internal/repository"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type URLService struct {
	urlRepo   repository.URLRepository
	cacheRepo repository.CacheRepository
}

func NewURLService(urlRepo repository.URLRepository, cacheRepo repository.CacheRepository) *URLService {
	return &URLService{urlRepo, cacheRepo}
}

func validateURL(urlStr string) error {
	if strings.TrimSpace(urlStr) == "" {
		return errors.New("URL cannot be empty")
	}

	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return errors.New("URL must start with http:// or https://")
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return errors.New("invalid URL format")
	}

	if parsed.Hostname() == "" {
		return errors.New("URL must contain a hostname")
	}

	if len(parsed.Hostname()) > 253 {
		return errors.New("hostname is too long")
	}

	return nil
}

func randomString(l int) string {
	b := make([]byte, l)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (s *URLService) CreateShortURL(ctx context.Context, originalURL string) (*models.URL, error) {
	if err := validateURL(originalURL); err != nil {
		return nil, err
	}

	existing, err := s.GetURLByOriginal(ctx, originalURL)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	var shortCode string
	maxAttempts := 10

	for i := 0; i < maxAttempts; i++ {
		shortCode = randomString(6)
		exists, err := s.urlRepo.FindByShortCode(ctx, shortCode)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if exists == nil {
			break
		}
		if i == maxAttempts-1 {
			return nil, errors.New("failed to generate unique short code")
		}
	}

	newURL := &models.URL{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
	}

	if err := s.urlRepo.Create(ctx, newURL); err != nil {
		return nil, err
	}

	if err := s.cacheRepo.SetURL(ctx, newURL); err != nil {
		return nil, err
	}

	return newURL, nil
}

func (s *URLService) GetURL(ctx context.Context, shortCode string) (*models.URL, error) {
	url, err := s.cacheRepo.GetURL(ctx, shortCode)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	if url != nil {
		return url, nil
	}

	url, err = s.urlRepo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	if err := s.cacheRepo.SetURL(ctx, url); err != nil {
		log.Printf("Failed to cache URL: %v", err)
	}

	return url, nil
}

func (s *URLService) GetURLByOriginal(ctx context.Context, originalURL string) (*models.URL, error) {
	return s.urlRepo.FindByOriginalURL(ctx, originalURL)
}
