package authservice

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Да, я знаю, что без хранения у меня не будет возможности отзывать токе
// и тд и тд, но уже у меня просто нету на это времени
type AuthService struct {
	tokenDuration time.Duration
	secret        []byte
}

func (a *AuthService) GenerateToken(userID string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(a.tokenDuration).Unix()
	claims["iat"] = time.Now().Unix()

	tokenString, err := token.SignedString(a.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (a *AuthService) ParseToken(token string) (string, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.secret, nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		userID := claims["user_id"].(string)
		return userID, nil
	}

	return "", ErrInvalidToken
}

func (a *AuthService) VerifyToken(token string) error {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.secret, nil
	})
	if err != nil {
		return a.errorValidationHandler(err)
	}

	if !parsedToken.Valid {
		return ErrInvalidToken
	}

	return nil
}

func (a *AuthService) errorValidationHandler(err error) error {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return ErrExpiredToken
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		return ErrTokenNotActive
	case errors.Is(err, jwt.ErrTokenMalformed):
		return fmt.Errorf("token is malformed")
	default:
		return ErrInvalidToken
	}
}
