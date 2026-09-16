package jwt

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenValidityTime = 24 * time.Hour

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret []byte
}

func New(secret string) *Manager {
	return &Manager{
		secret: []byte(secret),
	}
}

func (m *Manager) Generate(ctx context.Context, id int64, role string) (string, error) {
	claims := &Claims{
		ID:   strconv.FormatInt(id, 10),
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenValidityTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString(m.secret)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (m *Manager) Validate(ctx context.Context, tokenStr string) (int64, string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return 0, fmt.Errorf("unexpected signing method: %v", token.Method)
		}
		return m.secret, nil
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
