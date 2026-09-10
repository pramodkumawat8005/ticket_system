package main

import (
	"errors"
	"fmt"
	"time"
	"crypto/rand"
	"encoding/hex"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("id-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}


const bcryptCost = 12

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}


func VerifyPassword(password, storedHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	return err == nil
}


var errInvalidToken = errors.New("invalid token")

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (c Claims) UserID() string {
	return c.Subject
}

func GenerateToken(secret []byte, userID, username string, ttl time.Duration) (string, error) {
	if len(secret) < 16 {
		return "", errors.New("JWT secret must be at least 16 bytes")
	}
	if ttl <= 0 {
		return "", errors.New("token TTL must be positive")
	}

	now := time.Now().UTC()
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}


func ParseToken(secret []byte, tokenString string) (*Claims, error) {
	if len(secret) < 16 {
		return nil, errInvalidToken
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errInvalidToken
	}
	if claims.UserID() == "" || claims.Username == "" {
		return nil, errInvalidToken
	}

	return claims, nil
}
