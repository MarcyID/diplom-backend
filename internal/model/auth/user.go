package auth

import "time"

// User представляет пользователя в системе
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Никогда не возвращать в JSON
	FullName  *string   `json:"full_name,omitempty"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	BannerURL *string   `json:"banner_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest - запрос на регистрацию
type CreateUserRequest struct {
	Email    string  `json:"email" binding:"required,email"`
	Username string  `json:"username" binding:"required,min=3,max=50"`
	Password string  `json:"password" binding:"required,min=6,max=72"`
	FullName *string `json:"full_name,omitempty"`
}

// LoginRequest - запрос на вход
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse - ответ при авторизации
type AuthResponse struct {
	User         UserInfo `json:"user"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
}

// UserInfo - публичная информация о пользователе
type UserInfo struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FullName  *string   `json:"full_name,omitempty"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	BannerURL *string   `json:"banner_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ToUserInfo конвертирует User в UserInfo
func (u *User) ToUserInfo() UserInfo {
	return UserInfo{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		FullName:  u.FullName,
		AvatarURL: u.AvatarURL,
		BannerURL: u.BannerURL,
		CreatedAt: u.CreatedAt,
	}
}
