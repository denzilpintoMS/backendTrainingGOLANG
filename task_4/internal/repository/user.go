package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"task_3/internal/model"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailInUse   = errors.New("email already registered")
)

type UserRepository interface {
	CreateUser(context.Context, model.User) (model.User, error)
	GetUserByEmail(context.Context, string) (model.User, error)
}

var _ UserRepository = (*Postgres)(nil)

func (r *Postgres) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.User{}, ErrEmailInUse
		}
		return model.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (r *Postgres) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("lower(email) = ?", strings.ToLower(strings.TrimSpace(email))).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}
