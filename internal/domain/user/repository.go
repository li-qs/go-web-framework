package user

import (
	"context"
	"myframework/ent"
	"myframework/ent/token"
	entUser "myframework/ent/user"
	"time"
)

type UserRepo struct {
	user *ent.UserClient
}

func (r *UserRepo) GetByID(ctx context.Context, userID int) (*ent.User, error) {
	return r.user.Get(ctx, userID)
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*ent.User, error) {
	return r.user.
		Query().
		Where(entUser.Username(username)).
		Only(ctx)
}

func (r *UserRepo) UpdatePasswordHash(ctx context.Context, userID int, passwordHash string) error {
	_, err := r.user.
		UpdateOneID(userID).
		SetPasswordHash(passwordHash).
		Save(ctx)
	return err
}

type TokenRepo struct {
	token *ent.TokenClient
}

func (r *TokenRepo) GetByToken(ctx context.Context, tokenHash string) (*ent.Token, error) {
	return r.token.
		Query().
		Where(token.TokenHash(tokenHash)).
		Only(ctx)
}

func (r *TokenRepo) Create(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	_, err := r.token.
		Create().
		SetUserID(userID).
		SetTokenHash(tokenHash).
		SetExpiresAt(expiresAt).
		Save(ctx)
	return err
}

func (r *TokenRepo) Delete(ctx context.Context, id int) error {
	return r.token.DeleteOneID(id).Exec(ctx)
}

func (r *TokenRepo) DeleteByUserID(ctx context.Context, userID int) error {
	_, err := r.token.
		Delete().
		Where(token.UserID(userID)).
		Exec(ctx)
	return err
}
