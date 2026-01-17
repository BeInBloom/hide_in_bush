package middlewares

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BeInBloom/hide_in_bush/internal/middlewares/mocks"
	"go.uber.org/mock/gomock"
)

func TestAuth_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockauthService(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mw := New(logger, mockAuth)

	mockAuth.EXPECT().VerifyToken("valid-token").Return(nil)
	mockAuth.EXPECT().ParseToken("valid-token").Return("user-123", nil)

	handler := mw.Auth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserIDFromContext(r.Context())
		if userID != "user-123" {
			t.Errorf("UserID in context = %v, want user-123", userID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "valid-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockauthService(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mw := New(logger, mockAuth)

	mockAuth.EXPECT().VerifyToken("invalid-token").Return(errors.New("invalid token"))

	handler := mw.Auth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "invalid-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuth_ParseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockauthService(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mw := New(logger, mockAuth)

	mockAuth.EXPECT().VerifyToken("token").Return(nil)
	mockAuth.EXPECT().ParseToken("token").Return("", errors.New("parse error"))

	handler := mw.Auth()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusUnauthorized)
	}
}
