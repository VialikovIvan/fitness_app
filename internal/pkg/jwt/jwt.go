package myjwt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	authTokenValidityTime   = 24 * time.Hour
	accessTokenValidityTime = 5 * time.Minute
)

var ErrInvalidToken = errors.New("invalid token")

type AuthClaims struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type AccessClaims struct {
	ID string `json:"id"`
	jwt.RegisteredClaims
}

type Manager struct {
	authSecret   []byte
	accessSecret []byte
}

func New(authSecret string, accessSecret string) *Manager {
	return &Manager{
		authSecret:   []byte(authSecret),
		accessSecret: []byte(accessSecret),
	}
}

func (m *Manager) GenerateAuth(ctx context.Context, id int64, role string) (string, error) {
	claims := &AuthClaims{
		ID:   strconv.FormatInt(id, 10),
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(authTokenValidityTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString(m.authSecret)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (m *Manager) ValidateAuth(ctx context.Context, tokenStr string) (int64, string, error) {
	claims := &AuthClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return 0, fmt.Errorf("unexpected signing method: %v", token.Method)
		}
		return m.authSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, "", ErrInvalidToken
	}

	id, err := strconv.ParseInt(claims.ID, 10, 64)
	if err != nil {
		return 0, "", err
	}

	return id, claims.Role, nil
}

func (m *Manager) GenerateAccess(ctx context.Context, id int64) (string, error) {
	b := make([]byte, 16)
	rand.Read(b)

	claims := &AccessClaims{
		ID: strconv.FormatInt(id, 10),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        hex.EncodeToString(b),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenValidityTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString(m.accessSecret)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (m *Manager) ValidateAccess(ctx context.Context, tokenStr string) (int64, string, error) {
	claims := &AccessClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return 0, fmt.Errorf("unexpected signing method: %v", token.Method)
		}
		return m.accessSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, "", ErrInvalidToken
	}

	id, err := strconv.ParseInt(claims.ID, 10, 64)
	if err != nil {
		return 0, "", err
	}

	return id, claims.RegisteredClaims.ID, nil
}
