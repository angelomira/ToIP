package repo

import (
	"context"
	"errors"

	"example.com/pz9-auth/internal/core"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailTaken   = errors.New("email already in use")
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) AutoMigrate() error {
	return r.db.AutoMigrate(&core.User{})
}

func (r *UserRepo) Create(ctx context.Context, user *core.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrEmailTaken
		}
		return err
	}
	return nil
}

func (r *UserRepo) ByEmail(ctx context.Context, email string) (core.User, error) {
	var user core.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return core.User{}, ErrUserNotFound
	}
	return user, err
}
