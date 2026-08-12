package user

import (
	"context"
	"time"
)

type UserRepoImpl interface {
	GetByID(ctx context.Context, userID int64) (*UserEntity, error)
	GetByUsername(ctx context.Context, username string) (*UserEntity, error)
	UpdatePasswordHash(ctx context.Context, userID int64, passwordHash string) error
}

type TokenRepoImpl interface {
	GetByToken(ctx context.Context, tokenHash string) (*TokenEntity, error)
	Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	Delete(ctx context.Context, userID int64) error
	DeleteByUserID(ctx context.Context, userID int64) error
}
