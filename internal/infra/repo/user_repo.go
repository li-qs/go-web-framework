package repo

import (
	"context"
	"myframework/internal/domain/user"
	"myframework/pkg/mysql"
)

type UserRepo struct {
	db *mysql.DB
}

func NewUserRepo(db *mysql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*user.UserEntity, error) {
	var user user.UserEntity
	err := r.db.Get(&user, "SELECT * FROM `user` WHERE id=?", id)
	return &user, err
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*user.UserEntity, error) {
	var user user.UserEntity
	err := r.db.Get(&user, "SELECT * FROM `user` WHERE username=?", username)
	return &user, err
}

func (r *UserRepo) UpdatePasswordHash(ctx context.Context, id int64, passwordHash string) error {
	_, err := r.db.Exec("UPDATE `user` SET password_hash=? WHERE id=?", passwordHash, id)
	return err
}
