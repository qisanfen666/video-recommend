package service

import (
	"errors"
	"time"
	"video-recommend/internal/domain"
	"video-recommend/internal/repository"
	"video-recommend/pkg/jwt"
	"video-recommend/pkg/utils"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
	}
}

func (s *UserService) Register(req *domain.UserRegisterRequest) (*domain.User, error) {
	existing, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	if req.Email != "" {
		existing, err = s.userRepo.GetByEmail(req.Email)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, errors.New("email already registered")
		}
	}

	user := &domain.User{
		Username: req.Username,
		Password: utils.MD5(req.Password),
		Email:    req.Email,
		Role:     "user",
		Gender:   "unknown",
		Nickname: req.Username,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(req *domain.UserLoginRequest) (*domain.UserLoginResponse, error) {
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if user.Password != utils.MD5(req.Password) {
		return nil, ErrInvalidCredentials
	}

	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user.LastLoginAt = &now
	s.userRepo.UpdateLastLogin(user.ID)

	return &domain.UserLoginResponse{
		Token:     token,
		ExpiresIn: int64(24 * 3600),
		User:      user,
	}, nil
}

func (s *UserService) GetByID(id int64) (*domain.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) UpdateProfile(userID int64, req *domain.UserProfileRequest) (*domain.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Age > 0 {
		user.Age = req.Age
	}
	if req.Gender != "" {
		user.Gender = req.Gender
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Preferences != "" {
		user.Preferences = req.Preferences
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) List(page, pageSize int) ([]domain.User, int64, error) {
	return s.userRepo.List(page, pageSize)
}

func (s *UserService) GetPreferences(userID int64) ([]string, error) {
	return s.userRepo.GetPreferences(userID)
}
