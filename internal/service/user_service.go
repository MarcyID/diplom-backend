package service

import (
	"context"
	"diplomM/internal/model/auth"
	"diplomM/internal/repository"
)

// UserService предоставляет бизнес-логику для работы с пользователями
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService создает новый UserService
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetProfile получает профиль пользователя
func (s *UserService) GetProfile(ctx context.Context, userID int64) (*auth.UserInfo, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	userInfo := user.ToUserInfo()
	return &userInfo, nil
}

// UpdateGenrePreferences обновляет жанровые предпочтения пользователя
func (s *UserService) UpdateGenrePreferences(ctx context.Context, userID int64, genreIDs []int64) error {
	return s.userRepo.UpdateGenrePreferences(ctx, userID, genreIDs)
}

// GetGenrePreferences получает жанровые предпочтения пользователя
func (s *UserService) GetGenrePreferences(ctx context.Context, userID int64) ([]int64, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Возвращаем пустой срез вместо nil, если предпочтения не установлены
	if user.GenrePreferences == nil {
		return []int64{}, nil
	}

	return user.GenrePreferences, nil
}
