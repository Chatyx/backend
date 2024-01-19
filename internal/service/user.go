package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Chatyx/backend/internal/dto"
	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/hasher"
)

type UserRepository interface {
	List(ctx context.Context) ([]entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id int) (entity.User, error)
	GetByUsername(ctx context.Context, username string) (entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	UpdatePassword(ctx context.Context, userID int, pwdHash string) error
	Delete(ctx context.Context, id int) error
}

type SessionRepository interface {
	DeleteAllByUserID(ctx context.Context, id string) error
}

type UserConfig struct {
	UserRepository    UserRepository
	SessionRepository SessionRepository
}

type User struct {
	userRepo UserRepository
	sessRepo SessionRepository
	hasher   hasher.BCrypt
}

func NewUser(conf UserConfig) *User {
	return &User{
		userRepo: conf.UserRepository,
		sessRepo: conf.SessionRepository,
	}
}

func (u *User) List(ctx context.Context) ([]entity.User, error) {
	users, err := u.userRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list of users: %w", err)
	}

	return users, nil
}

func (u *User) Create(ctx context.Context, obj dto.UserCreate) (entity.User, error) {
	pwdHash, err := u.hasher.Hash(obj.Password)
	if err != nil {
		return entity.User{}, fmt.Errorf("get hash from password: %w", err)
	}

	user := entity.User{
		Username:  obj.Username,
		PwdHash:   pwdHash,
		Email:     obj.Email,
		FirstName: obj.FirstName,
		LastName:  obj.LastName,
		BirthDate: obj.BirthDate,
		Bio:       obj.Bio,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	if err = u.userRepo.Create(ctx, &user); err != nil {
		return entity.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (u *User) GetByID(ctx context.Context, id int) (entity.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return entity.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (u *User) Update(ctx context.Context, obj dto.UserUpdate) (entity.User, error) {
	user := entity.User{
		ID:        obj.ID,
		Username:  obj.Username,
		Email:     obj.Email,
		FirstName: obj.FirstName,
		LastName:  obj.LastName,
		BirthDate: obj.BirthDate,
		Bio:       obj.Bio,
		IsActive:  true,
		UpdatedAt: time.Now(),
	}

	if err := u.userRepo.Update(ctx, &user); err != nil {
		return entity.User{}, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

func (u *User) UpdatePassword(ctx context.Context, obj dto.UserUpdatePassword) error {
	user, err := u.userRepo.GetByID(ctx, obj.UserID) // TODO: think about for update (lock)
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}

	if !u.hasher.CompareHashAndPassword(user.PwdHash, obj.CurPassword) {
		return fmt.Errorf("compare current hash and password: %w", entity.ErrWrongCurrentPassword)
	}

	pwdHash, err := u.hasher.Hash(obj.NewPassword)
	if err != nil {
		return fmt.Errorf("get hash from new password: %w", err)
	}

	if err = u.userRepo.UpdatePassword(ctx, user.ID, pwdHash); err != nil {
		return fmt.Errorf("update new password: %w", err)
	}

	return nil
}

func (u *User) CheckPassword(ctx context.Context, username, password string) (userID string, ok bool, err error) {
	user, err := u.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("get user by name: %w", err)
	}

	return strconv.Itoa(user.ID), u.hasher.CompareHashAndPassword(user.PwdHash, password), nil
}

func (u *User) Delete(ctx context.Context, id int) error {
	if err := u.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if err := u.sessRepo.DeleteAllByUserID(ctx, strconv.Itoa(id)); err != nil {
		return fmt.Errorf("delete all user sessions: %w", err)
	}

	return nil
}
