package repo

import (
	"context"
	"myframework/internal/domain/user"
	"myframework/pkg/mysql"
	"time"
)

type TokenRepo struct {
	db            *mysql.DB
	expireSeconds int32
}

func NewTokenRepo(db *mysql.DB, expireSeconds int32) *TokenRepo {
	return &TokenRepo{db: db, expireSeconds: expireSeconds}
}

func (r *TokenRepo) Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		"INSERT INTO `refresh_token` (user_id, token_hash, expires_at) VALUES (?, ?, FROM_UNIXTIME(?))",
		userID, tokenHash, expiresAt,
	)
	return err
}

func (r *TokenRepo) GetByToken(ctx context.Context, tokenHash string) (*user.TokenEntity, error) {
	var rt user.TokenEntity
	err := r.db.Get(&rt, "SELECT * FROM `refresh_token` WHERE token_hash=?", tokenHash)
	return &rt, err
}

func (r *TokenRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec("DELETE FROM `refresh_token` WHERE id=?", id)
	return err
}

func (r *TokenRepo) DeleteByUserID(ctx context.Context, userID int64) error {
	_, err := r.db.Exec("DELETE FROM `refresh_token` WHERE user_id=?", userID)
	return err
}
