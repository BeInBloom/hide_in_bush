package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BeInBloom/hide_in_bush/internal/handlers/mocks"
	"github.com/BeInBloom/hide_in_bush/internal/models"
	"github.com/BeInBloom/hide_in_bush/internal/storage"
	"go.uber.org/mock/gomock"
)

func setupHandlers(t *testing.T) (*Handlers, *mocks.MockuserService, *mocks.MockauthService, *mocks.MockorderService, *mocks.MockwithdrawalService, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockUserService := mocks.NewMockuserService(ctrl)
	mockAuthService := mocks.NewMockauthService(ctrl)
	mockOrderService := mocks.NewMockorderService(ctrl)
	mockWithdrawalService := mocks.NewMockwithdrawalService(ctrl)
	h := New(mockUserService, mockAuthService, mockOrderService, mockWithdrawalService)
	return h, mockUserService, mockAuthService, mockOrderService, mockWithdrawalService, ctrl
}

func TestRegisterUserHandler_200_Success(t *testing.T) {
	h, mockUserService, mockAuthService, _, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().
		Register(gomock.Any(), models.UserCredentials{Login: "testuser", Password: "password123"}).
		Return("user-123", nil)
	mockAuthService.EXPECT().
		GenerateToken("user-123").
		Return("jwt-token", nil)

	body := `{"login":"testuser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	h.RegisterUserHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
	}
	if rr.Header().Get("Authorization") != "jwt-token" {
		t.Errorf("Authorization = %v, want jwt-token", rr.Header().Get("Authorization"))
	}
}

func TestRegisterUserHandler_400_InvalidJSON(t *testing.T) {
	h, _, _, _, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	h.RegisterUserHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusBadRequest)
	}
}

func TestRegisterUserHandler_409_UserExists(t *testing.T) {
	h, mockUserService, _, _, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().
		Register(gomock.Any(), gomock.Any()).
		Return("", storage.ErrUserAlreadyExists)

	body := `{"login":"existinguser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	h.RegisterUserHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusConflict)
	}
}

func TestRegisterUserHandler_500_InternalError(t *testing.T) {
	h, mockUserService, _, _, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().
		Register(gomock.Any(), gomock.Any()).
		Return("", errors.New("database error"))

	body := `{"login":"testuser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	h.RegisterUserHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusInternalServerError)
	}
}

func TestLoginUserHandler_200_Success(t *testing.T) {
	h, mockUserService, mockAuthService, _, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().
		ValidateCredentials(gomock.Any(), models.UserCredentials{Login: "testuser", Password: "password123"}).
		Return("user-123", nil)
	mockAuthService.EXPECT().
		GenerateToken("user-123").
		Return("jwt-token", nil)

	body := `{"login":"testuser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	h.LoginUserHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
	}
}

func TestLoginUserHandler_400_InvalidJSON(t *testing.T) {
	h, _, _, _, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	body := `{bad json`
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	h.LoginUserHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusBadRequest)
	}
}

func TestLoginUserHandler_401_WrongPassword(t *testing.T) {
	h, mockUserService, _, _, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().
		ValidateCredentials(gomock.Any(), gomock.Any()).
		Return("", errors.New("invalid credentials"))

	body := `{"login":"testuser","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	h.LoginUserHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusBadRequest)
	}
}

func TestLoginUserHandler_404_UserNotFound(t *testing.T) {
	h, mockUserService, _, _, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockUserService.EXPECT().
		ValidateCredentials(gomock.Any(), gomock.Any()).
		Return("", storage.ErrUserNotFound)

	body := `{"login":"nonexistent","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	h.LoginUserHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusNotFound)
	}
}

func TestUploadOrderHandler_202_Accepted(t *testing.T) {
	h, _, mockAuthService, mockOrderService, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)
	mockOrderService.EXPECT().UploadOrder(gomock.Any(), gomock.Any()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("12345678903"))
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.UploadOrderHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusAccepted)
	}
}

func TestUploadOrderHandler_200_AlreadyExists(t *testing.T) {
	h, _, mockAuthService, mockOrderService, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)
	mockOrderService.EXPECT().UploadOrder(gomock.Any(), gomock.Any()).Return(storage.ErrOrderAlreadyRegistered)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("12345678903"))
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.UploadOrderHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
	}
}

func TestUploadOrderHandler_409_OtherUser(t *testing.T) {
	h, _, mockAuthService, mockOrderService, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)
	mockOrderService.EXPECT().UploadOrder(gomock.Any(), gomock.Any()).Return(storage.ErrOrderRegisteredToOtherUser)

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("12345678903"))
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.UploadOrderHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusConflict)
	}
}

func TestUploadOrderHandler_500_InternalError(t *testing.T) {
	h, _, mockAuthService, mockOrderService, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)
	mockOrderService.EXPECT().UploadOrder(gomock.Any(), gomock.Any()).Return(errors.New("db error"))

	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("12345678903"))
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.UploadOrderHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusInternalServerError)
	}
}

func TestGetUserOrdersHandler_200_Success(t *testing.T) {
	h, _, mockAuthService, mockOrderService, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)
	mockOrderService.EXPECT().GetUserOrders(gomock.Any(), "user-123").Return([]models.Order{
		{ID: "9278923470", Status: "PROCESSED", Accrual: 500, Uploaded: time.Now()},
		{ID: "12345678903", Status: "PROCESSING", Uploaded: time.Now()},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.GetUserOrdersHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
	}
}

func TestGetUserOrdersHandler_204_NoOrders(t *testing.T) {
	h, _, mockAuthService, mockOrderService, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)
	mockOrderService.EXPECT().GetUserOrders(gomock.Any(), "user-123").Return([]models.Order{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.GetUserOrdersHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusNoContent)
	}
}

func TestGetUserOrdersHandler_500_InternalError(t *testing.T) {
	h, _, mockAuthService, mockOrderService, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)
	mockOrderService.EXPECT().GetUserOrders(gomock.Any(), "user-123").Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.GetUserOrdersHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusInternalServerError)
	}
}

func TestGetUserBalanceHandler_200_Success(t *testing.T) {
	h, mockUserService, mockAuthService, _, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)
	mockUserService.EXPECT().UserBalance(gomock.Any(), "user-123").Return(models.Balance{CurrentBalance: 500.5, Withdrawn: 42.0}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.GetUserBalanceHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
	}

	var balance models.Balance
	json.NewDecoder(rr.Body).Decode(&balance)
	if balance.CurrentBalance != 500.5 {
		t.Errorf("CurrentBalance = %v, want 500.5", balance.CurrentBalance)
	}
	if balance.Withdrawn != 42.0 {
		t.Errorf("Withdrawn = %v, want 42.0", balance.Withdrawn)
	}
}

func TestGetUserBalanceHandler_500_InternalError(t *testing.T) {
	h, mockUserService, mockAuthService, _, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)
	mockUserService.EXPECT().UserBalance(gomock.Any(), "user-123").Return(models.Balance{}, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.GetUserBalanceHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusInternalServerError)
	}
}

func TestWithdrawPointsHandler_200_Success(t *testing.T) {
	h, _, mockAuthService, _, mockWithdrawalService, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)
	mockWithdrawalService.EXPECT().PostWithdraw(gomock.Any(), "user-123", gomock.Any()).Return(nil)

	body := `{"order":"2377225624","sum":751}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.WithdrawPointsHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
	}
}

func TestWithdrawPointsHandler_400_InvalidJSON(t *testing.T) {
	h, _, mockAuthService, _, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)

	body := `{bad json`
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.WithdrawPointsHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusBadRequest)
	}
}

func TestWithdrawPointsHandler_401_Unauthorized(t *testing.T) {
	h, _, mockAuthService, _, _, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("invalid-token").Return("", errors.New("invalid token"))

	body := `{"order":"2377225624","sum":751}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "invalid-token")
	rr := httptest.NewRecorder()

	h.WithdrawPointsHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusUnauthorized)
	}
}

func TestWithdrawPointsHandler_402_InsufficientBalance(t *testing.T) {
	h, _, mockAuthService, _, mockWithdrawalService, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)
	mockWithdrawalService.EXPECT().PostWithdraw(gomock.Any(), "user-123", gomock.Any()).Return(storage.ErrInsufficientBalance)

	body := `{"order":"2377225624","sum":999999}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.WithdrawPointsHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusPaymentRequired {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusPaymentRequired)
	}
}

func TestWithdrawPointsHandler_500_InternalError(t *testing.T) {
	h, _, mockAuthService, _, mockWithdrawalService, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)
	mockWithdrawalService.EXPECT().PostWithdraw(gomock.Any(), "user-123", gomock.Any()).Return(errors.New("db error"))

	body := `{"order":"2377225624","sum":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.WithdrawPointsHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusInternalServerError)
	}
}

func TestGetWithdrawalsHandler_200_Success(t *testing.T) {
	h, _, mockAuthService, _, mockWithdrawalService, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)
	mockWithdrawalService.EXPECT().GetUserWithdrawals(gomock.Any(), "user-123").Return([]models.Withdrawal{
		{Order: "2377225624", Sum: 500, ProcessedAt: time.Now()},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.GetWithdrawalsHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusOK)
	}
}

func TestGetWithdrawalsHandler_204_NoWithdrawals(t *testing.T) {
	h, _, mockAuthService, _, mockWithdrawalService, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)
	mockWithdrawalService.EXPECT().GetUserWithdrawals(gomock.Any(), "user-123").Return([]models.Withdrawal{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.GetWithdrawalsHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusNoContent)
	}
}

func TestGetWithdrawalsHandler_500_InternalError(t *testing.T) {
	h, _, mockAuthService, _, mockWithdrawalService, ctrl := setupHandlers(t)
	defer ctrl.Finish()

	mockAuthService.EXPECT().ParseToken("jwt-token").Return("user-123", nil)
	mockWithdrawalService.EXPECT().GetUserWithdrawals(gomock.Any(), "user-123").Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
	req.Header.Set("Authorization", "jwt-token")
	rr := httptest.NewRecorder()

	h.GetWithdrawalsHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %v, want %v", rr.Code, http.StatusInternalServerError)
	}
}
