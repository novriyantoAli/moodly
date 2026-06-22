package service

import (
	"context"
	"errors"
	"time"

	"github.com/novriyantoAli/moodly/internal/application/user/dto"
	"github.com/novriyantoAli/moodly/internal/application/user/entity"
	"github.com/novriyantoAli/moodly/internal/application/user/repository"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	ValidateUser(ctx context.Context, id uint) (*dto.UserResponse, error)
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*dto.UserResponse, error)
	GetUsers(ctx context.Context, filter *dto.UserFilter) (*dto.UserListResponse, error)
	UpdateUser(ctx context.Context, id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, id uint) error
	Login(ctx context.Context, req *dto.LoginUserRequest) (*dto.LoginUserResponse, error)
}

type userService struct {
	repo   repository.UserRepository
	logger *zap.Logger
}

func NewUserService(repo repository.UserRepository, logger *zap.Logger) UserService {
	return &userService{
		repo:   repo,
		logger: logger,
	}
}

func (s *userService) ValidateUser(ctx context.Context, id uint) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	return s.entityToResponse(user), nil
}

func (s *userService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	exists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		s.logger.Error("Failed to check email existence", zap.Error(err))
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, errors.New("failed to process password")
	}

	user := &entity.User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		FullName:  req.FullName,
		Level:     "user",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.repo.Create(ctx, user)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(user), nil
}

func (s *userService) GetUserByID(ctx context.Context, id uint) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return s.entityToResponse(user), nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*dto.UserResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return s.entityToResponse(user), nil
}

func (s *userService) GetUsers(ctx context.Context, filter *dto.UserFilter) (*dto.UserListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	users, totalCount, err := s.repo.GetAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, *s.entityToResponse(&user))
	}

	return &dto.UserListResponse{
		Data:       responses,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
}

func (s *userService) UpdateUser(ctx context.Context, id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Level != "" {
		user.Level = req.Level
	}
	user.IsActive = req.IsActive
	user.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, user)
	if err != nil {
		s.logger.Error("Failed to update user", zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(user), nil
}

func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *userService) Login(ctx context.Context, req *dto.LoginUserRequest) (*dto.LoginUserResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Verify password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		s.logger.Info("Invalid password attempt", zap.String("email", req.Email))
		return nil, errors.New("invalid email or password")
	}

	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	return &dto.LoginUserResponse{
		ID:       user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		Level:    user.Level,
	}, nil
}

func (s *userService) entityToResponse(user *entity.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		Level:     user.Level,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
