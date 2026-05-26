package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"diplomM/internal/model/auth"
	"diplomM/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTClaims кастомные claims для JWT токенов
type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthServiceConfig конфигурация сервиса аутентификации
type AuthServiceConfig struct {
	JWTSecretKey    string
	AccessTokenExp  time.Duration // Время жизни access токена (например, 15 минут)
	RefreshTokenExp time.Duration // Время жизни refresh токена (например, 7 дней)
}

// AuthService сервис аутентификации
type AuthService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	config      AuthServiceConfig
}

// NewAuthService создает новый AuthService
func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	config AuthServiceConfig,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		config:      config,
	}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, req auth.CreateUserRequest) (*auth.User, error) {
	// Проверяем, существует ли пользователь с таким email
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Проверяем, существует ли пользователь с таким username
	existingUser, err = s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this username already exists")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &auth.User{
		Email:    req.Email,
		Username: req.Username,
		Password: string(hashedPassword),
		FullName: req.FullName,
	}

	// Создаем пользователя в БД
	return s.userRepo.Create(ctx, user)
}

// Login выполняет вход пользователя
func (s *AuthService) Login(ctx context.Context, req auth.LoginRequest, userAgent *string, ipAddress *string) (*auth.AuthResponse, error) {
	// Находим пользователя по email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Генерируем токены
	accessToken, refreshToken, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Создаем сессию
	session, err := s.createSession(user.ID, refreshToken, userAgent, ipAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	_ = session // session.ID можно использовать для логирования

	return &auth.AuthResponse{
		User:         user.ToUserInfo(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Logout выполняет выход пользователя (удаляет сессию)
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	// Хэшируем refresh токен для поиска в БД
	tokenHash := s.hashToken(refreshToken)

	// Находим сессию
	session, err := s.sessionRepo.GetByToken(ctx, tokenHash)
	if err != nil {
		// Сессия не найдена или истекла - это не ошибка
		return nil
	}

	// Удаляем сессию
	return s.sessionRepo.Delete(ctx, session.ID)
}

// GetMe возвращает информацию о текущем пользователе
func (s *AuthService) GetMe(ctx context.Context, userID int64) (*auth.UserInfo, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	userInfo := user.ToUserInfo()
	return &userInfo, nil
}

// RefreshToken обновляет access и refresh токены
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*auth.AuthResponse, error) {
	// Хэшируем refresh токен для поиска в БД
	tokenHash := s.hashToken(refreshToken)

	// Находим сессию
	session, err := s.sessionRepo.GetByToken(ctx, tokenHash)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Проверяем, не истекла ли сессия
	if session.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("refresh token expired")
	}

	// Получаем пользователя
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Генерируем новые токены
	newAccessToken, newRefreshToken, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Создаем новую сессию (старую можно удалить)
	_ = s.sessionRepo.Delete(ctx, session.ID)

	_, err = s.createSession(user.ID, newRefreshToken, session.UserAgent, session.IPAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create new session: %w", err)
	}

	return &auth.AuthResponse{
		User:         user.ToUserInfo(),
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// generateTokens генерирует пару access и refresh токенов
func (s *AuthService) generateTokens(user *auth.User) (string, string, error) {
	now := time.Now()

	// Access токен
	accessClaims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.AccessTokenExp)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWTSecretKey))
	if err != nil {
		return "", "", err
	}

	// Refresh токен (просто случайная строка)
	refreshClaims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.config.RefreshTokenExp)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.config.JWTSecretKey))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// createSession создает сессию в БД
func (s *AuthService) createSession(userID int64, refreshToken string, userAgent *string, ipAddress *string) (*auth.Session, error) {
	tokenHash := s.hashToken(refreshToken)

	session := &auth.Session{
		UserID:       userID,
		RefreshToken: tokenHash,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    time.Now().Add(s.config.RefreshTokenExp),
	}

	return s.sessionRepo.Create(context.Background(), session)
}

// hashToken хэширует токен для безопасного хранения
func (s *AuthService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// ValidateToken проверяет валидность JWT токена
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// GetUserByID получает пользователя по ID
func (s *AuthService) GetUserByID(ctx context.Context, id int64) (*auth.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// UpdateUser обновляет пользователя
func (s *AuthService) UpdateUser(ctx context.Context, user *auth.User) error {
	return s.userRepo.Update(ctx, user)
}
