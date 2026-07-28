package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"myapi/internal/model"
	"myapi/internal/repository"
	"myapi/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

var algorithm = sha256.New
var invalidTokenErr = fmt.Errorf("invalid token")

type Login struct {
	UserRepo           *repository.User
	LoginExpireSeconds int
}

func (s *Login) AuthUser(username, password string) (*model.User, bool, error) {
	user, err := s.UserRepo.GetByUsername(username)
	if err != nil {
		return nil, false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return user, err == nil, nil
}

func (s *Login) generateTokenHash(user *model.User, ts int64, nonce string) string {
	str := fmt.Sprintf("%s%s%d%s", user.Username, nonce, ts, user.PasswordHash)
	b := sha256.Sum256([]byte(str))
	hash := hex.EncodeToString(b[:])
	return hash
}

func (s *Login) GenerateToken(user *model.User) (string, int, error) {
	ts := time.Now().UnixMilli()
	nonce, err := utils.GenerateRandomString(32)
	if err != nil {
		return "", 0, err
	}
	hash := s.generateTokenHash(user, ts, nonce)
	username := utils.Base64URLEncode(user.Username)
	return fmt.Sprintf("%s.%s.%d.%s", username, nonce, ts, hash), s.LoginExpireSeconds, nil
}

func (s *Login) AuthToken(token string) (*model.User, bool, error) {
	a := strings.Split(token, ".")
	if len(a) != 4 {
		return nil, false, invalidTokenErr
	}

	username, err := utils.Base64URLDecode(a[0])
	if err != nil {
		return nil, false, invalidTokenErr
	}

	nonce := a[1]

	ts, err := strconv.Atoi(a[2])
	if err != nil {
		return nil, false, invalidTokenErr
	}

	hash := a[3]

	now := time.Now().UnixMilli()
	const maxSkew = int64(5 * time.Second / time.Millisecond)
	if int64(ts)-now > maxSkew || now-int64(ts) > int64(s.LoginExpireSeconds)*1000 {
		return nil, false, invalidTokenErr
	}

	user, err := s.UserRepo.GetByUsername(username)
	if err != nil {
		return nil, false, err
	}

	h := s.generateTokenHash(user, int64(ts), nonce)
	if h != hash {
		return nil, false, nil
	}
	return user, true, nil
}
