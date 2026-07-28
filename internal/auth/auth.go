// Package auth реализует хеширование паролей и JWT-токены.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const tokenTTL = 24 * time.Hour

// Claims - JWT claims с идентификатором пользователя.
type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"uid"`
}

// Service предоставляет операции аутентификации.
type Service struct {
	secret []byte
}

// NewService создаёт сервис с заданным секретом подписи JWT.
// Если секрет пустой, генерируется случайный и логируется предупреждение.
func NewService(secret string) *Service {
	if secret == "" {
		// Генерируем случайный секрет для разработки
		// В продакшене JWT_SECRET должен быть задан явно
		randomSecret := make([]byte, 32)
		if _, err := rand.Read(randomSecret); err != nil {
			panic(fmt.Sprintf("failed to generate random JWT secret: %v", err))
		}
		secret = hex.EncodeToString(randomSecret)
		slog.Warn("JWT_SECRET not set, using random secret (not suitable for production)")
	}
	return &Service{secret: []byte(secret)}
}

// HashPassword хеширует пароль с помощью bcrypt.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword проверяет соответствие пароля хешу.
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// GenerateToken выпускает JWT для пользователя.
func (s *Service) GenerateToken(userID uuid.UUID) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// ValidateToken проверяет JWT и возвращает ID пользователя.
func (s *Service) ValidateToken(tokenString string) (uuid.UUID, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}
	return claims.UserID, nil
}
