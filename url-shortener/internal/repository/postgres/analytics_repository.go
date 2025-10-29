package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"url-shortener/internal/models"
)

type PostgresClickRepo struct {
	db *sql.DB
}

func NewPostgresClickRepo(db *sql.DB) *PostgresClickRepo {
	return &PostgresClickRepo{db: db}
}

func (p *PostgresClickRepo) SaveClick(ctx context.Context, click *models.Click) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {

		}
	}(tx)

	if click.CreatedAt.IsZero() {
		click.CreatedAt = time.Now()
	}

	query := `INSERT INTO clicks (id, url_id, ip_address, user_agent, referer, created_at)
              VALUES ($1, $2, $3, $4, $5, $6)`

	err = tx.QueryRowContext(
		ctx,
		query,
		click.ID,
		click.URLID,
		click.IPAddress,
		click.UserAgent,
		click.Referer,
		click.CreatedAt,
	).Scan(&click.ID)

	if err != nil {
		return fmt.Errorf("failed to insert click: %w", err)
	}

	UpdatedAt := time.Now()

	query = `UPDATE urls SET click_count = click_count + 1, updated_at = $1 WHERE id = $2`

	result, err := tx.ExecContext(ctx, query, UpdatedAt, click.URLID)
	if err != nil {
		return fmt.Errorf("failed to update click count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no URL found with id %s", click.URLID)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p *PostgresClickRepo) GetAnalyticsByID(ctx context.Context, ID int) (*models.Analytics, error) {
	var a models.Analytics

	query := `SELECT COUNT(*) as click_count FROM clicks WHERE url_id = $1 AND created_at > NOW() - INTERVAL '1 week'`

	var totalClicks int
	err := p.db.QueryRowContext(ctx, query, ID).Scan(&totalClicks)
	if err != nil {
		return nil, err
	}
	a.TotalClicks = totalClicks

	query = `SELECT 
				TO_CHAR(created_at, 'YYYY-MM-DD') as day_date,
				COUNT(*) as click_daily_count
			FROM clicks 
			WHERE url_id = $1 AND created_at >= CURRENT_DATE - INTERVAL '6 days'
			GROUP BY TO_CHAR(created_at, 'YYYY-MM-DD')
			ORDER BY day_date DESC`

	rows, err := p.db.QueryContext(ctx, query, ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var dailyClick models.DailyClick
		err := rows.Scan(&dailyClick.Date, &dailyClick.Count)
		if err != nil {
			return nil, err
		}
		a.DailyClicks = append(a.DailyClicks, dailyClick)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	query = `SELECT 
				referer,
				COUNT(*) as referrer_count
			FROM clicks 
			WHERE url_id = $1
			GROUP BY referer
			ORDER BY referrer_count DESC`

	rows, err = p.db.QueryContext(ctx, query, ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var referrerStat models.ReferrerStat
		err := rows.Scan(&referrerStat.Referrer, &referrerStat.Count)
		if err != nil {
			return nil, err
		}
		if totalClicks > 0 {
			percent := float64(referrerStat.Count) * 100 / float64(totalClicks)
			referrerStat.Percent = fmt.Sprintf("%.1f%%", percent)
		} else {
			referrerStat.Percent = "0%"
		}
		a.Referrers = append(a.Referrers, referrerStat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	query = `SELECT 
				user_agent,
				COUNT(*) as browser_count
			FROM clicks 
			WHERE url_id = $1
			GROUP BY user_agent
			ORDER BY browser_count DESC`

	rows, err = p.db.QueryContext(ctx, query, ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var browserStat models.BrowserStat
		err := rows.Scan(&browserStat.Browser, &browserStat.Count)
		if err != nil {
			return nil, err
		}
		a.Browsers = append(a.Browsers, browserStat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *PostgresClickRepo) GetAnalyticsByShortCode(ctx context.Context, shortCode string) (*models.Analytics, error) {
	var urlID int
	query := `SELECT id FROM urls WHERE short_code = $1`
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(&urlID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find URL by short code: %w", err)
	}

	return r.GetAnalyticsByID(ctx, urlID)
}
