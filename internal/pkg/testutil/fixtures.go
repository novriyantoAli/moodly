package testutil

import (
	"time"

	"github.com/google/uuid"
	authEntity "github.com/novriyantoAli/moodly/internal/application/auth/entity"
	billDto "github.com/novriyantoAli/moodly/internal/application/bill/dto"
	billEntity "github.com/novriyantoAli/moodly/internal/application/bill/entity"
	"github.com/novriyantoAli/moodly/internal/application/payment/dto"
	"github.com/novriyantoAli/moodly/internal/application/payment/entity"
	securityEntity "github.com/novriyantoAli/moodly/internal/application/security/entity"
	subscribeDto "github.com/novriyantoAli/moodly/internal/application/subscribe/dto"
	subscribeEntity "github.com/novriyantoAli/moodly/internal/application/subscribe/entity"
	userDto "github.com/novriyantoAli/moodly/internal/application/user/dto"
	userEntity "github.com/novriyantoAli/moodly/internal/application/user/entity"
)

func CreateLoginAttemptFixture() *authEntity.LoginAttempt {
	uid := uint(1)
	return &authEntity.LoginAttempt{
		UserID:    &uid,
		Username:  "testuser",
		Success:   false,
		CreatedAt: time.Now(),
	}
}

func CreateAuthSessionFixture() *authEntity.AuthSession {
	return &authEntity.AuthSession{
		UserID:       1,
		AccessToken:  "access_token_value",
		RefreshToken: "refresh_token_value",
		ExpiredAt:    time.Now().Add(24 * time.Hour),
	}
}

func CreateUserPasswordFixture() *securityEntity.UserPassword {
	return &securityEntity.UserPassword{
		UserID:        1,
		PasswordHash:  "hashed_password_value",
		FailedAttempt: 0,
		LockedUntil:   nil,
	}
}

func CreateUserPINFixture() *securityEntity.UserPIN {
	return &securityEntity.UserPIN{
		UserID:        1,
		PinHash:       "hashed_pin_value",
		FailedAttempt: 0,
		LockedUntil:   nil,
	}
}

// User fixtures
func CreateUserFixture() *userEntity.User {
	return &userEntity.User{
		ID:        1,
		Email:     "john@example.com",
		FullName:  "John Doe",
		Level:     "user",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func CreateUserRequestFixture() *userDto.CreateUserRequest {
	return &userDto.CreateUserRequest{
		Email:    "john@example.com",
		Password: "password123",
		FullName: "John Doe",
	}
}

func CreateUpdateUserRequestFixture() *userDto.UpdateUserRequest {
	return &userDto.UpdateUserRequest{
		FullName: "John Updated",
		Level:    "user",
		IsActive: true,
	}
}

// Payment fixtures
func CreatePaymentFixture() *entity.Payment {
	billID := uuid.Must(uuid.NewRandom())
	return &entity.Payment{
		ID:            1,
		PaymentNumber: uuid.New().String(),
		Amount:        10050,
		Currency:      "USD",
		Status:        entity.PaymentStatusPending,
		Description:   "Test payment",
		BillID:        billID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func CreatePaymentRequestFixture() *dto.CreatePaymentRequest {
	return &dto.CreatePaymentRequest{
		Method:     "cash",
		BillNumber: "123456789",
	}
}

func CreateUpdatePaymentRequestFixture() *dto.UpdatePaymentRequest {
	return &dto.UpdatePaymentRequest{
		Status:      entity.PaymentStatusCompleted.String(),
		Description: "Payment completed",
	}
}

func CreatePaymentFilterFixture() *dto.PaymentFilter {
	billID := uuid.Must(uuid.NewRandom())
	return &dto.PaymentFilter{
		Status:   "pending",
		Currency: "USD",
		BillID:   billID.String(),
		Page:     1,
		PageSize: 10,
	}
}

// Subscriber fixtures
func CreateSubscriberFixture() *subscribeEntity.Subscriber {
	return &subscribeEntity.Subscriber{
		ID:        1,
		Username:  "testuser",
		CallName:  "Test User",
		Password:  "password123",
		Plan:      "pppoe",
		Price:     50000.0,
		StartDate: time.Now(),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func CreateSubscriberRequestFixture() *subscribeDto.CreateSubscriberRequest {
	return &subscribeDto.CreateSubscriberRequest{
		Username:  "testuser",
		CallName:  "Test User",
		Password:  "password123",
		Plan:      "pppoe",
		Price:     50000.0,
		StartDate: time.Now(),
	}
}

func CreateUpdateSubscriberRequestFixture() *subscribeDto.UpdateSubscriberRequest {
	return &subscribeDto.UpdateSubscriberRequest{
		CallName: "Updated User",
		Plan:     "hotspot",
		Price:    30000.0,
		IsActive: true,
	}
}

func CreateSubscriberFilterFixture() *subscribeDto.SubscribeFilter {
	return &subscribeDto.SubscribeFilter{
		Page:     1,
		PageSize: 10,
	}
}

// Bill fixtures
func CreateBillFixture() *billEntity.Bill {
	return &billEntity.Bill{
		ID:          uuid.New(),
		BillNumber:  uuid.New().String(),
		SubscribeID: 1,
		BillMonth:   3,
		BillYear:    2026,
		Amount:      150000,
		DueDate:     time.Now().AddDate(0, 1, 0),
		Status:      "unpaid",
		CreatedAt:   time.Now(),
	}
}

func CreateBillRequestFixture() *billDto.CreateBillRequest {
	return &billDto.CreateBillRequest{
		SubscribeID: 1,
		BillMonth:   3,
		BillYear:    2026,
		Amount:      150000,
		DueDate:     time.Now().AddDate(0, 1, 0),
		Status:      "unpaid",
	}
}

func CreateUpdateBillRequestFixture() *billDto.UpdateBillRequest {
	return &billDto.UpdateBillRequest{
		Amount:  150000,
		DueDate: time.Now().AddDate(0, 1, 0),
		Status:  "paid",
	}
}

func CreateBillFilterFixture() *billDto.BillFilter {
	return &billDto.BillFilter{
		SubscribeID: 1,
		Status:      "unpaid",
		Page:        1,
		PageSize:    10,
	}
}

func CreateUserOAuthFixture() *securityEntity.UserOAuth {
	return &securityEntity.UserOAuth{
		Provider:       "google",
		ProviderUserID: "provider-user-id",
		Email:          "user@example.com",
		Name:           "Test User",
		Picture:        "https://example.com/avatar.jpg",
	}
}
