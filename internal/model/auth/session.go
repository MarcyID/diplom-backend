package auth

import "time"

// Session представляет сессию пользователя (для refresh токенов)
type Session struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	RefreshToken string    `json:"-"` // Не возвращать в JSON
	UserAgent    *string   `json:"user_agent,omitempty"`
	IPAddress    *string   `json:"ip_address,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	IsRevoked    bool      `json:"-"`
}

// CreateSessionRequest - запрос на создание сессии
type CreateSessionRequest struct {
	UserID       int64
	RefreshToken string
	UserAgent    *string
	IPAddress    *string
	ExpiresAt    time.Time
}
