package authservice

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken(t *testing.T) {
	service := New()

	token, err := service.GenerateToken("user123")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}
}

func TestParseToken_Valid(t *testing.T) {
	service := New()

	userID := "test-user-id"
	token, _ := service.GenerateToken(userID)

	parsedUserID, err := service.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	if parsedUserID != userID {
		t.Errorf("ParseToken() = %v, want %v", parsedUserID, userID)
	}
}

func TestParseToken_Invalid(t *testing.T) {
	service := New()

	_, err := service.ParseToken("invalid.token.here")
	if err == nil {
		t.Error("ParseToken() expected error for invalid token")
	}
}

func TestVerifyToken_Valid(t *testing.T) {
	service := New()

	token, _ := service.GenerateToken("user123")

	err := service.VerifyToken(token)
	if err != nil {
		t.Errorf("VerifyToken() error = %v", err)
	}
}

func TestVerifyToken_Invalid(t *testing.T) {
	service := New()

	err := service.VerifyToken("invalid.token")
	if err == nil {
		t.Error("VerifyToken() expected error for invalid token")
	}
}

func TestVerifyToken_Expired(t *testing.T) {
	service := &AuthService{
		tokenDuration: -1 * time.Hour,
		secret:        []byte(secret),
	}

	token, _ := service.GenerateToken("user123")

	err := service.VerifyToken(token)
	if err == nil {
		t.Error("VerifyToken() expected error for expired token")
	}
}

func TestParseToken_WrongSigningMethod(t *testing.T) {
	service := New()

	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"user_id": "user123",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	_, err := service.ParseToken(tokenString)
	if err == nil {
		t.Error("ParseToken() expected error for wrong signing method")
	}
}
