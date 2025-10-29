package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"url-shortener/internal/models"
)

type PostgresURLRepo struct {
	db *sql.DB
}

func NewPostgresURLRepo(db *sql.DB) *PostgresURLRepo {
	return &PostgresURLRepo{db: db}
}

func (p *PostgresURLRepo) Create(ctx context.Context, url *models.URL) error {
	if url.CreatedAt.IsZero() {
		url.CreatedAt = time.Now()
	}

	if url.UpdatedAt.IsZero() {
		url.UpdatedAt = time.Now()
	}

	query := `
		INSERT INTO urls (original_url, short_code, created_at, updated_at, click_count)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := p.db.QueryRowContext(
		ctx,
		query,
		url.OriginalURL,
		url.ShortCode,
		url.CreatedAt,
		url.UpdatedAt,
		url.ClickCount,
	).Scan(&url.ID)

	if err != nil {
		return fmt.Errorf("failed to insert URL: %w", err)
	}

	return nil
}

func (p *PostgresClickRepo) IncrementClickCount(ctx context.Context, ID int) error {
	query := `UPDATE urls SET click_count = click_count + 1 WHERE id = $1`

	row, err := p.db.ExecContext(ctx, query, ID)
	if err != nil {
		return fmt.Errorf("failed to update click count: %w", err)
	}
	rowCount, err := row.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rowCount == 0 {
		return fmt.Errorf("no click found with id %d", ID)
	}
	return nil
}

func (p *PostgresURLRepo) GetByID(ctx context.Context, ID int) (*models.URL, error) {
	query := `SELECT id, original_url, short_code, created_at, updated_at, click_count FROM urls WHERE id = $1`

	row := p.db.QueryRowContext(ctx, query, ID)

	var url models.URL
	if err := row.Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.ClickCount,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to scan task: %w", err)
	}

	return &url, nil
}

func (p *PostgresURLRepo) FindByShortCode(ctx context.Context, shortCode string) (*models.URL, error) {
	query := `SELECT id, original_url, short_code, created_at, updated_at, click_count FROM urls WHERE short_code = $1`

	row := p.db.QueryRowContext(ctx, query, shortCode)
	var url models.URL
	if err := row.Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.ClickCount,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to scan URL: %w", err)
	}

	return &url, nil
}

func (p *PostgresURLRepo) Update(ctx context.Context, url *models.URL) error {
	url.UpdatedAt = time.Now()

	query := `UPDATE urls SET original_url = $1, short_code = $2, updated_at = $3, click_count = $4 WHERE id = $5`

	result, err := p.db.ExecContext(
		ctx,
		query,
		url.OriginalURL,
		url.ShortCode,
		url.UpdatedAt,
		url.ClickCount,
		url.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update URL: %w", err)
	}

	rowsAffect, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rowsAffect == 0 {
		return fmt.Errorf("task with id %d not found", url.ID)
	}

	return nil
}

func (p *PostgresURLRepo) Delete(ctx context.Context, ID int) error {
	query := `DELETE FROM urls WHERE id = $1`

	row, err := p.db.ExecContext(ctx, query, ID)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}

	rowsAffect, err := row.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rowsAffect == 0 {
		return fmt.Errorf("task with id %d not found", ID)
	}
	return nil
}

func (p *PostgresURLRepo) FindByOriginalURL(ctx context.Context, originalURL string) (*models.URL, error) {
	query := `SELECT id, original_url, short_code, created_at, updated_at, click_count 
              FROM urls WHERE original_url = $1`

	var url models.URL
	err := p.db.QueryRowContext(ctx, query, originalURL).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.ClickCount,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find URL by original: %w", err)
	}

	return &url, nil
}
