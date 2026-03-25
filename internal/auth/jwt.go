package auth

import (
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type JWTService struct {
	mu          sync.RWMutex
	secret      []byte
	expireHours int
}

func NewJWTService(secret string, expireHours int) *JWTService {
	return &JWTService{
		secret:      []byte(secret),
		expireHours: expireHours,
	}
}

func (s *JWTService) UpdateSecret(secret string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secret = []byte(secret)
}

func (s *JWTService) CurrentSecret() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return string(s.secret)
}

func (s *JWTService) GenerateToken(userID uint, username string) (string, error) {
	s.mu.RLock()
	secret := append([]byte(nil), s.secret...)
	expireHours := s.expireHours
	s.mu.RUnlock()

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (s *JWTService) ParseToken(tokenString string) (*Claims, error) {
	s.mu.RLock()
	secret := append([]byte(nil), s.secret...)
	s.mu.RUnlock()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
