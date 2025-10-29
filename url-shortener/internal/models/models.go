package models

import (
	"time"
)

// URL представляет основную сущность - сокращенную ссылку
type URL struct {
	ID          int       `json:"id" db:"id"`
	OriginalURL string    `json:"original_url" db:"original_url"`
	ShortCode   string    `json:"short_code" db:"short_code"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	ClickCount  int       `json:"click_count" db:"click_count"`
}

// Click представляет запись о каждом переходе по короткой ссылке
type Click struct {
	ID        int       `json:"id" db:"id"`
	URLID     int       `json:"url_id" db:"url_id"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	Referer   string    `json:"referer" db:"referer"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Analytics содержит агрегированную статистику по кликам
type Analytics struct {
	TotalClicks int            `json:"total_clicks"` // Общее количество кликов
	DailyClicks []DailyClick   `json:"daily_clicks"` // Статистика по дням
	Referrers   []ReferrerStat `json:"referrers"`    // Статистика по источникам переходов
	Browsers    []BrowserStat  `json:"browsers"`     // Статистика по браузерам
}

// DailyClick представляет количество кликов за конкретный день
type DailyClick struct {
	Date  string `json:"date" db:"date"`   // Дата в формате YYYY-MM-DD
	Count int    `json:"count" db:"count"` // Количество кликов за эту дату
}

// ReferrerStat представляет статистику по доменам-источникам переходов
type ReferrerStat struct {
	Referrer string `json:"referrer"` // Домен источника (google.com, direct, facebook.com)
	Count    int    `json:"count"`    // Количество переходов с этого источника
	Percent  string `json:"percent"`  // Процент от общего числа (например "15.5%")
}

// BrowserStat представляет статистику по браузерам пользователей
type BrowserStat struct {
	Browser string `json:"browser"` // Название браузера (Chrome, Firefox, Safari)
	Version string `json:"version"` // Версия браузера ("91.0", "89.0")
	Count   int    `json:"count"`   // Количество переходов с этого браузера
}
