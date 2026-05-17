package repository

import (
	"errors"
	"video-recommend/internal/domain"
	"video-recommend/pkg/database"

	"gorm.io/gorm"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(user *domain.User) error {
	return database.DB.Create(user).Error
}

func (r *UserRepository) GetByID(id int64) (*domain.User, error) {
	var user domain.User
	err := database.DB.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(username string) (*domain.User, error) {
	var user domain.User
	err := database.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := database.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *domain.User) error {
	return database.DB.Save(user).Error
}

func (r *UserRepository) Delete(id int64) error {
	return database.DB.Delete(&domain.User{}, id).Error
}

func (r *UserRepository) List(page, pageSize int) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	query := database.DB.Model(&domain.User{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id desc").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) UpdateLastLogin(id int64) error {
	return database.DB.Model(&domain.User{}).Where("id = ?", id).Update("last_login_at", gorm.Expr("NOW()")).Error
}

func (r *UserRepository) GetPreferences(userID int64) ([]string, error) {
	user, err := r.GetByID(userID)
	if err != nil || user == nil {
		return nil, err
	}
	return parsePreferences(user.Preferences), nil
}

func parsePreferences(pref string) []string {
	if pref == "" {
		return []string{}
	}
	var prefs []string
	for _, p := range splitAndTrim(pref, ",") {
		if p != "" {
			prefs = append(prefs, p)
		}
	}
	return prefs
}

func splitAndTrim(s, sep string) []string {
	var result []string
	var current string
	for _, c := range s {
		if string(c) == sep {
			if current != "" {
				result = append(result, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
