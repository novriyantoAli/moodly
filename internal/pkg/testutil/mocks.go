package testutil

import (
	"context"
	"time"

	"github.com/google/uuid"
	authDto "github.com/novriyantoAli/moodly/internal/application/auth/dto"
	authEntity "github.com/novriyantoAli/moodly/internal/application/auth/entity"
	billDto "github.com/novriyantoAli/moodly/internal/application/bill/dto"
	billEntity "github.com/novriyantoAli/moodly/internal/application/bill/entity"
	"github.com/novriyantoAli/moodly/internal/application/payment/dto"
	"github.com/novriyantoAli/moodly/internal/application/payment/entity"
	securityDto "github.com/novriyantoAli/moodly/internal/application/security/dto"
	securityEntity "github.com/novriyantoAli/moodly/internal/application/security/entity"
	subscribeDto "github.com/novriyantoAli/moodly/internal/application/subscribe/dto"
	subscribeEntity "github.com/novriyantoAli/moodly/internal/application/subscribe/entity"
	userDto "github.com/novriyantoAli/moodly/internal/application/user/dto"
	userEntity "github.com/novriyantoAli/moodly/internal/application/user/entity"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/idtoken"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *userEntity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uint) (*userEntity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userEntity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*userEntity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userEntity.User), args.Error(1)
}

func (m *MockUserRepository) GetAll(ctx context.Context, filter *userDto.UserFilter) ([]userEntity.User, int64, error) {
	args := m.Called(ctx, filter)
	var users []userEntity.User
	if args.Get(0) != nil {
		users = args.Get(0).([]userEntity.User)
	}

	var count int64
	if args.Get(1) != nil {
		count = args.Get(1).(int64)
	}
	return users, count, args.Error(2)
}

func (m *MockUserRepository) Update(ctx context.Context, user *userEntity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GetUsersByRoleName(ctx context.Context, roleName string, filter *userDto.UserFilter) ([]userEntity.User, int64, error) {
	args := m.Called(ctx, roleName, filter)
	var users []userEntity.User
	if args.Get(0) != nil {
		users = args.Get(0).([]userEntity.User)
	}

	var count int64
	if args.Get(1) != nil {
		count = args.Get(1).(int64)
	}
	return users, count, args.Error(2)
}

// MockPaymentRepository is a mock implementation of PaymentRepository
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(
	ctx context.Context,
	payment *entity.Payment,
) error {

	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByID(
	ctx context.Context,
	id uint,
) (*entity.Payment, error) {

	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entity.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetByIDForUpdate(
	ctx context.Context,
	id uint,
) (*entity.Payment, error) {

	args := m.Called(
		ctx,
		id,
	)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entity.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetByPaymentNumber(
	ctx context.Context,
	paymentNumber string,
) (*entity.Payment, error) {

	args := m.Called(ctx, paymentNumber)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entity.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetByBillID(
	ctx context.Context,
	billID string,
) ([]entity.Payment, error) {

	args := m.Called(ctx, billID)

	var payments []entity.Payment
	if args.Get(0) != nil {
		payments = args.Get(0).([]entity.Payment)
	}

	return payments, args.Error(1)
}

func (m *MockPaymentRepository) GetAll(
	ctx context.Context,
	filter *dto.PaymentFilter,
) ([]entity.Payment, int64, error) {

	args := m.Called(ctx, filter)

	var payments []entity.Payment
	if args.Get(0) != nil {
		payments = args.Get(0).([]entity.Payment)
	}

	var totalCount int64
	if args.Get(1) != nil {
		totalCount = args.Get(1).(int64)
	}

	return payments, totalCount, args.Error(2)
}

func (m *MockPaymentRepository) Update(
	ctx context.Context,
	payment *entity.Payment,
) error {

	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) ExistsActivePaymentByBillID(
	ctx context.Context,
	billID string,
) (bool, error) {

	args := m.Called(ctx, billID)

	return args.Bool(0), args.Error(1)
}

type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) GetPaymentsByBillID(
	ctx context.Context,
	billID string,
) ([]dto.PaymentResponse, error) {

	args := m.Called(ctx, billID)

	var payments []dto.PaymentResponse
	if args.Get(0) != nil {
		payments = args.Get(0).([]dto.PaymentResponse)
	}

	return payments, args.Error(1)
}

func (m *MockPaymentService) GetPaymentByNumber(
	ctx context.Context,
	paymentNumber string,
) (*dto.PaymentResponse, error) {

	args := m.Called(ctx, paymentNumber)

	var payment *dto.PaymentResponse
	if args.Get(0) != nil {
		payment = args.Get(0).(*dto.PaymentResponse)
	}

	return payment, args.Error(1)
}

func (m *MockPaymentService) CreatePayment(ctx context.Context, req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentResponse), args.Error(1)
}

func (m *MockPaymentService) GetPaymentByID(ctx context.Context, id uint) (*dto.PaymentResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentResponse), args.Error(1)
}

func (m *MockPaymentService) GetPayments(ctx context.Context, filter *dto.PaymentFilter) (*dto.PaymentListResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentListResponse), args.Error(1)
}

func (m *MockPaymentService) UpdatePayment(ctx context.Context, id uint, req *dto.UpdatePaymentRequest) (*dto.PaymentResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PaymentResponse), args.Error(1)
}

func (m *MockPaymentService) DeletePayment(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPaymentService) GetPaymentsByUser(ctx context.Context, userID uint) ([]dto.PaymentResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.PaymentResponse), args.Error(1)
}

// MockSubscribeRepository is a mock implementation of SubscribeRepository
type MockSubscribeRepository struct {
	mock.Mock
}

func (m *MockSubscribeRepository) Create(ctx context.Context, subscriber *subscribeEntity.Subscriber) error {
	args := m.Called(ctx, subscriber)
	return args.Error(0)
}

func (m *MockSubscribeRepository) GetByID(ctx context.Context, id uint) (*subscribeEntity.Subscriber, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscribeEntity.Subscriber), args.Error(1)
}

func (m *MockSubscribeRepository) GetByUsername(ctx context.Context, username string) (*subscribeEntity.Subscriber, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscribeEntity.Subscriber), args.Error(1)
}

func (m *MockSubscribeRepository) GetActiveSubscribes(ctx context.Context) ([]subscribeEntity.Subscriber, error) {
	args := m.Called(ctx)
	var subscribers []subscribeEntity.Subscriber
	if args.Get(0) != nil {
		subscribers = args.Get(0).([]subscribeEntity.Subscriber)
	}
	return subscribers, args.Error(1)
}

func (m *MockSubscribeRepository) GetAll(ctx context.Context, filter *subscribeDto.SubscribeFilter) ([]subscribeEntity.Subscriber, int64, error) {
	args := m.Called(ctx, filter)
	var subscribers []subscribeEntity.Subscriber
	if args.Get(0) != nil {
		subscribers = args.Get(0).([]subscribeEntity.Subscriber)
	}

	var count int64
	if args.Get(1) != nil {
		count = args.Get(1).(int64)
	}
	return subscribers, count, args.Error(2)
}

func (m *MockSubscribeRepository) Update(ctx context.Context, subscriber *subscribeEntity.Subscriber) error {
	args := m.Called(ctx, subscriber)
	return args.Error(0)
}

func (m *MockSubscribeRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSubscribeRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockSubscribeRepository) Count(ctx context.Context, filter *subscribeDto.CountFilter) (int64, error) {
	args := m.Called(ctx, filter)
	var count int64
	if args.Get(0) != nil {
		count = args.Get(0).(int64)
	}
	return count, args.Error(1)
}

// MockSubscribeService is a mock implementation of SubscribeService
type MockSubscribeService struct {
	mock.Mock
}

func (m *MockSubscribeService) CreateSubscriber(ctx context.Context, req *subscribeDto.CreateSubscriberRequest) (*subscribeDto.SubscriberResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscribeDto.SubscriberResponse), args.Error(1)
}

func (m *MockSubscribeService) GetSubscriberByID(ctx context.Context, id uint) (*subscribeDto.SubscriberResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscribeDto.SubscriberResponse), args.Error(1)
}

func (m *MockSubscribeService) GetSubscriberByUsername(ctx context.Context, username string) (*subscribeDto.SubscriberResponse, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscribeDto.SubscriberResponse), args.Error(1)
}

func (m *MockSubscribeService) GetSubscribers(ctx context.Context, filter *subscribeDto.SubscribeFilter) (*subscribeDto.SubscriberListResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscribeDto.SubscriberListResponse), args.Error(1)
}

func (m *MockSubscribeService) UpdateSubscriber(ctx context.Context, id uint, req *subscribeDto.UpdateSubscriberRequest) (*subscribeDto.SubscriberResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscribeDto.SubscriberResponse), args.Error(1)
}

func (m *MockSubscribeService) DeleteSubscriber(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSubscribeService) CountFilter(ctx context.Context, filter *subscribeDto.CountFilter) (*subscribeDto.CountResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*subscribeDto.CountResponse), args.Error(1)
}

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) ValidateUser(
	ctx context.Context,
	id uint,
) (*userDto.UserResponse, error) {

	args := m.Called(ctx, id)

	var user *userDto.UserResponse
	if args.Get(0) != nil {
		user = args.Get(0).(*userDto.UserResponse)
	}

	return user, args.Error(1)
}

func (m *MockUserService) CreateUser(ctx context.Context, req *userDto.CreateUserRequest) (*userDto.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id uint) (*userDto.UserResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*userDto.UserResponse, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUsers(ctx context.Context, filter *userDto.UserFilter) (*userDto.UserListResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserListResponse), args.Error(1)
}

func (m *MockUserService) GetPsikologUsers(ctx context.Context, filter *userDto.UserFilter) (*userDto.UserListResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserListResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id uint, req *userDto.UpdateUserRequest) (*userDto.UserResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserService) Login(ctx context.Context, req *userDto.LoginUserRequest) (*userDto.LoginUserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.LoginUserResponse), args.Error(1)
}

// MockBillRepository is a mock implementation of BillRepository
type MockBillRepository struct {
	mock.Mock
}

func (m *MockBillRepository) Create(ctx context.Context, bill *billEntity.Bill) error {
	args := m.Called(ctx, bill)
	return args.Error(0)
}

func (m *MockBillRepository) GetByNumber(
	ctx context.Context,
	billNumber string,
) (*billEntity.Bill, error) {

	args := m.Called(ctx, billNumber)

	var bill *billEntity.Bill
	if args.Get(0) != nil {
		bill = args.Get(0).(*billEntity.Bill)
	}

	return bill, args.Error(1)
}

func (m *MockBillRepository) LockByNumber(
	ctx context.Context,
	billNumber string,
) (*billEntity.Bill, error) {

	args := m.Called(ctx, billNumber)

	var bill *billEntity.Bill
	if args.Get(0) != nil {
		bill = args.Get(0).(*billEntity.Bill)
	}

	return bill, args.Error(1)
}

func (m *MockBillRepository) GetByID(ctx context.Context, id uuid.UUID) (*billEntity.Bill, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*billEntity.Bill), args.Error(1)
}

func (m *MockBillRepository) GetAll(ctx context.Context, filter *billDto.BillFilter) ([]billEntity.Bill, int64, error) {
	args := m.Called(ctx, filter)
	var bills []billEntity.Bill
	if args.Get(0) != nil {
		bills = args.Get(0).([]billEntity.Bill)
	}

	var count int64
	if args.Get(1) != nil {
		count = args.Get(1).(int64)
	}
	return bills, count, args.Error(2)
}

func (m *MockBillRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockBillRepository) Exists(ctx context.Context, subscribeID uint, month uint, year uint) (bool, error) {
	args := m.Called(ctx, subscribeID, month, year)
	return args.Bool(0), args.Error(1)
}

func (m *MockBillRepository) CountUnpaidBills(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBillRepository) SumUnpaidBillsAmount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBillRepository) CountOverdueBills(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBillRepository) SumOverdueBillsAmount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBillRepository) GetActiveSubWithUnpaidBills(ctx context.Context) ([]billEntity.Bill, error) {
	args := m.Called(ctx)
	var bills []billEntity.Bill
	if args.Get(0) != nil {
		bills = args.Get(0).([]billEntity.Bill)
	}
	return bills, args.Error(1)
}

func (m *MockBillRepository) SumAmountByMonthYear(ctx context.Context, status string, month uint, year uint) (int64, error) {
	args := m.Called(ctx, status, month, year)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBillRepository) CountBillsByMonthYear(ctx context.Context, filter *billDto.CountBillFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

// MockBillService is a mock implementation of BillService
type MockBillService struct {
	mock.Mock
}

func (m *MockBillService) ValidateForPayment(
	ctx context.Context,
	billNumber string,
) (*billDto.BillResponse, error) {

	args := m.Called(ctx, billNumber)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*billDto.BillResponse), args.Error(1)
}

func (m *MockBillService) LockForPayment(
	ctx context.Context,
	billNumber string,
) (*billDto.BillResponse, error) {

	args := m.Called(
		ctx,
		billNumber,
	)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*billDto.BillResponse), args.Error(1)
}

func (m *MockBillService) MarkAsPaid(
	ctx context.Context,
	id string,
) error {

	args := m.Called(ctx, id)

	return args.Error(0)
}

func (m *MockBillService) GetBillByNumber(
	ctx context.Context,
	billNumber string,
) (*billDto.BillResponse, error) {

	args := m.Called(ctx, billNumber)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*billDto.BillResponse), args.Error(1)
}

func (m *MockBillService) CreateBill(
	ctx context.Context,
	req *billDto.CreateBillRequest,
) (*billDto.BillResponse, error) {

	args := m.Called(ctx, req)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*billDto.BillResponse), args.Error(1)
}

func (m *MockBillService) GenerateMonthlyBills(
	ctx context.Context,
	month uint,
	year uint,
) error {

	args := m.Called(ctx, month, year)
	return args.Error(0)
}

func (m *MockBillService) GetBillByID(
	ctx context.Context,
	id string,
) (*billDto.BillResponse, error) {

	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*billDto.BillResponse), args.Error(1)
}

func (m *MockBillService) GetBills(
	ctx context.Context,
	filter *billDto.BillFilter,
) (*billDto.BillListResponse, error) {

	args := m.Called(ctx, filter)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*billDto.BillListResponse), args.Error(1)
}

func (m *MockBillService) UpdateBillStatus(
	ctx context.Context,
	id string,
	status string,
) error {

	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockBillService) SumAmountByMonthYear(
	ctx context.Context,
	filter *billDto.SumAmountBillFilter,
) (int64, error) {

	args := m.Called(ctx, filter)

	var amount int64
	if args.Get(0) != nil {
		amount = args.Get(0).(int64)
	}

	return amount, args.Error(1)
}

func (m *MockBillService) CountBillsByMonthYear(
	ctx context.Context,
	filter *billDto.CountBillFilter,
) (*billDto.CountBillResponse, error) {

	args := m.Called(ctx, filter)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*billDto.CountBillResponse), args.Error(1)
}

func (m *MockBillService) ProcessUpdateBillStatusFromUnpaidToOverdue(
	ctx context.Context,
	month uint,
	year uint,
) error {

	args := m.Called(ctx, month, year)
	return args.Error(0)
}

func (m *MockBillService) QuickCountUnpaidBills(
	ctx context.Context,
) (*billDto.BillQuickCountUnpaidResponse, error) {

	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*billDto.BillQuickCountUnpaidResponse), args.Error(1)
}

func (m *MockBillService) QuickCountOverdueBills(
	ctx context.Context,
) (*billDto.BillQuickCountOverdueResponse, error) {

	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*billDto.BillQuickCountOverdueResponse), args.Error(1)
}

// MockBillPublisher is a mock implementation of BillPublisher
type MockBillPublisher struct {
	mock.Mock
}

func (m *MockBillPublisher) ScheduleBillPerSubscribe(req billDto.CreateBillRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockBillPublisher) ScheduleBillPerSubscribeChangeFromUnpaidOverdue(req billEntity.Bill) error {
	args := m.Called(req)
	return args.Error(0)
}

// MockUserPINRepository is a mock implementation of UserPINRepository
type MockUserPINRepository struct {
	mock.Mock
}

func (m *MockUserPINRepository) GetByUserID(ctx context.Context, userID uint) (*securityEntity.UserPIN, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*securityEntity.UserPIN), args.Error(1)
}

func (m *MockUserPINRepository) Create(ctx context.Context, pin *securityEntity.UserPIN) error {
	args := m.Called(ctx, pin)
	return args.Error(0)
}

func (m *MockUserPINRepository) UpdatePIN(ctx context.Context, userID uint, pinHash string) error {
	args := m.Called(ctx, userID, pinHash)
	return args.Error(0)
}

func (m *MockUserPINRepository) IncrementFailedAttempt(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserPINRepository) ResetFailedAttempt(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserPINRepository) LockAccount(ctx context.Context, userID uint, duration time.Duration) error {
	args := m.Called(ctx, userID, duration)
	return args.Error(0)
}

func (m *MockUserPINRepository) Unlock(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserPINRepository) Delete(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockUserPINService is a mock implementation of UserPINService
type MockUserPINService struct {
	mock.Mock
}

func (m *MockUserPINService) SetPIN(ctx context.Context, req *securityDto.SetPINRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockUserPINService) VerifyPIN(ctx context.Context, req *securityDto.VerifyPINRequest) (bool, error) {
	args := m.Called(ctx, req)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserPINService) GetSecurity(ctx context.Context, userID uint) (*securityDto.UserPINResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*securityDto.UserPINResponse), args.Error(1)
}

func (m *MockUserPINService) IsAccountLocked(ctx context.Context, userID uint) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

// MockUserPasswordRepository is a mock implementation of UserPasswordRepository
type MockUserPasswordRepository struct {
	mock.Mock
}

func (m *MockUserPasswordRepository) GetByUserID(
	ctx context.Context,
	userID uint,
) (*securityEntity.UserPassword, error) {

	args := m.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*securityEntity.UserPassword), args.Error(1)
}

func (m *MockUserPasswordRepository) GetByUsername(
	ctx context.Context,
	username string,
) (*securityEntity.UserPassword, error) {

	args := m.Called(ctx, username)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*securityEntity.UserPassword), args.Error(1)
}

func (m *MockUserPasswordRepository) Create(
	ctx context.Context,
	password *securityEntity.UserPassword,
) error {

	args := m.Called(ctx, password)
	return args.Error(0)
}

func (m *MockUserPasswordRepository) UpdatePasswordHash(
	ctx context.Context,
	userID uint,
	passwordHash string,
) error {

	args := m.Called(ctx, userID, passwordHash)
	return args.Error(0)
}

func (m *MockUserPasswordRepository) IncrementFailedAttempt(
	ctx context.Context,
	userID uint,
) error {

	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserPasswordRepository) ResetFailedAttempt(
	ctx context.Context,
	userID uint,
) error {

	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserPasswordRepository) LockAccount(
	ctx context.Context,
	userID uint,
	duration time.Duration,
) error {

	args := m.Called(ctx, userID, duration)
	return args.Error(0)
}

func (m *MockUserPasswordRepository) Unlock(
	ctx context.Context,
	userID uint,
) error {

	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserPasswordRepository) UpdateLastLogin(
	ctx context.Context,
	userID uint,
	loginTime time.Time,
) error {

	args := m.Called(ctx, userID, loginTime)
	return args.Error(0)
}

func (m *MockUserPasswordRepository) Delete(
	ctx context.Context,
	userID uint,
) error {

	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockUserPasswordService is a mock implementation of UserPasswordService
type MockUserPasswordService struct {
	mock.Mock
}

func (m *MockUserPasswordService) SetPassword(
	ctx context.Context,
	req *securityDto.SetPasswordRequest,
) error {

	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockUserPasswordService) VerifyPassword(
	ctx context.Context,
	req *securityDto.VerifyPasswordRequest,
) (bool, error) {

	args := m.Called(ctx, req)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserPasswordService) ChangePassword(
	ctx context.Context,
	req *securityDto.ChangePasswordRequest,
) error {

	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockUserPasswordService) GetPasswordInfo(
	ctx context.Context,
	userID uint,
) (*securityDto.UserPasswordResponse, error) {

	args := m.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*securityDto.UserPasswordResponse), args.Error(1)
}

func (m *MockUserPasswordService) IsAccountLocked(
	ctx context.Context,
	userID uint,
) (bool, error) {

	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserPasswordService) DeletePassword(
	ctx context.Context,
	userID uint,
) error {

	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockAuthSessionRepository is a mock implementation of AuthSessionRepository
type MockAuthSessionRepository struct {
	mock.Mock
}

func (m *MockAuthSessionRepository) GetByID(
	ctx context.Context,
	id uint,
) (*authEntity.AuthSession, error) {

	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*authEntity.AuthSession), args.Error(1)
}

func (m *MockAuthSessionRepository) GetByRefreshToken(
	ctx context.Context,
	refreshToken string,
) (*authEntity.AuthSession, error) {

	args := m.Called(ctx, refreshToken)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*authEntity.AuthSession), args.Error(1)
}

func (m *MockAuthSessionRepository) GetByUserID(
	ctx context.Context,
	userID uint,
) ([]authEntity.AuthSession, error) {

	args := m.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]authEntity.AuthSession), args.Error(1)
}

func (m *MockAuthSessionRepository) Create(
	ctx context.Context,
	session *authEntity.AuthSession,
) error {

	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockAuthSessionRepository) UpdateAccessToken(
	ctx context.Context,
	id uint,
	accessToken string,
) error {

	args := m.Called(ctx, id, accessToken)
	return args.Error(0)
}

func (m *MockAuthSessionRepository) UpdateRefreshToken(
	ctx context.Context,
	id uint,
	refreshToken string,
	expiredAt time.Time,
) error {

	args := m.Called(
		ctx,
		id,
		refreshToken,
		expiredAt,
	)

	return args.Error(0)
}

func (m *MockAuthSessionRepository) Delete(
	ctx context.Context,
	id uint,
) error {

	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAuthSessionRepository) DeleteByRefreshToken(
	ctx context.Context,
	refreshToken string,
) error {

	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthSessionRepository) DeleteByUserID(
	ctx context.Context,
	userID uint,
) error {

	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthSessionRepository) DeleteExpiredSessions(
	ctx context.Context,
) error {

	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAuthSessionRepository) GetActiveSessionCount(
	ctx context.Context,
	userID uint,
) (int64, error) {

	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuthSessionRepository) GetOldestActiveSession(
	ctx context.Context,
	userID uint,
) (*authEntity.AuthSession, error) {

	args := m.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*authEntity.AuthSession), args.Error(1)
}

// MockAuthSessionService is a mock implementation of AuthSessionService
type MockAuthSessionService struct {
	mock.Mock
}

func (m *MockAuthSessionService) CreateSession(
	ctx context.Context,
	session *authEntity.AuthSession,
) error {

	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockAuthSessionService) GetSessionByID(
	ctx context.Context,
	id uint,
) (*authEntity.AuthSession, error) {

	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*authEntity.AuthSession), args.Error(1)
}

func (m *MockAuthSessionService) GetSessionByRefreshToken(
	ctx context.Context,
	refreshToken string,
) (*authEntity.AuthSession, error) {

	args := m.Called(ctx, refreshToken)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*authEntity.AuthSession), args.Error(1)
}

func (m *MockAuthSessionService) GetUserSessions(
	ctx context.Context,
	userID uint,
) ([]authEntity.AuthSession, error) {

	args := m.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]authEntity.AuthSession), args.Error(1)
}

func (m *MockAuthSessionService) RefreshSession(
	ctx context.Context,
	refreshToken string,
	newAccessToken string,
	newRefreshToken string,
	expiredAt time.Time,
) error {

	args := m.Called(
		ctx,
		refreshToken,
		newAccessToken,
		newRefreshToken,
		expiredAt,
	)

	return args.Error(0)
}

func (m *MockAuthSessionService) Logout(
	ctx context.Context,
	sessionID uint,
) error {

	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockAuthSessionService) LogoutByRefreshToken(
	ctx context.Context,
	refreshToken string,
) error {

	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthSessionService) LogoutAllUserSessions(
	ctx context.Context,
	userID uint,
) error {

	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthSessionService) DeleteExpiredSessions(
	ctx context.Context,
) error {

	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAuthSessionService) GetActiveSessionCount(
	ctx context.Context,
	userID uint,
) (int64, error) {

	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// MockLoginAttemptRepository is a mock implementation of LoginAttemptRepository
type MockLoginAttemptRepository struct {
	mock.Mock
}

func (m *MockLoginAttemptRepository) Create(
	ctx context.Context,
	attempt *authEntity.LoginAttempt,
) error {

	args := m.Called(ctx, attempt)
	return args.Error(0)
}

func (m *MockLoginAttemptRepository) GetByUserID(
	ctx context.Context,
	userID uint,
) ([]authEntity.LoginAttempt, error) {

	args := m.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]authEntity.LoginAttempt), args.Error(1)
}

func (m *MockLoginAttemptRepository) GetByUsername(
	ctx context.Context,
	username string,
) ([]authEntity.LoginAttempt, error) {

	args := m.Called(ctx, username)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]authEntity.LoginAttempt), args.Error(1)
}

// MockLoginAttemptService is a mock implementation of LoginAttemptService
type MockLoginAttemptService struct {
	mock.Mock
}

func (m *MockLoginAttemptService) CreateAttempt(
	ctx context.Context,
	attempt *authEntity.LoginAttempt,
) error {

	args := m.Called(ctx, attempt)
	return args.Error(0)
}

func (m *MockLoginAttemptService) GetAttemptsByUserID(
	ctx context.Context,
	userID uint,
) ([]authEntity.LoginAttempt, error) {

	args := m.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]authEntity.LoginAttempt), args.Error(1)
}

func (m *MockLoginAttemptService) GetAttemptsByUsername(
	ctx context.Context,
	username string,
) ([]authEntity.LoginAttempt, error) {

	args := m.Called(ctx, username)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]authEntity.LoginAttempt), args.Error(1)
}

func (m *MockLoginAttemptService) GetLatestAttemptByUserID(
	ctx context.Context,
	userID uint,
) (*authEntity.LoginAttempt, error) {

	args := m.Called(ctx, userID)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*authEntity.LoginAttempt), args.Error(1)
}

func (m *MockLoginAttemptService) GetLatestAttemptByUsername(
	ctx context.Context,
	username string,
) (*authEntity.LoginAttempt, error) {

	args := m.Called(ctx, username)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*authEntity.LoginAttempt), args.Error(1)
}

func (m *MockLoginAttemptService) GetFailedAttemptCountByUserID(
	ctx context.Context,
	userID uint,
) (int, error) {

	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockLoginAttemptService) GetFailedAttemptCountByUsername(
	ctx context.Context,
	username string,
) (int, error) {

	args := m.Called(ctx, username)
	return args.Int(0), args.Error(1)
}

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateToken(
	userID uint,
	email string,
	level string,
) (string, error) {

	args := m.Called(
		userID,
		email,
		level,
	)

	return args.String(0),
		args.Error(1)
}

func (m *MockJWTService) GenerateRefreshToken(
	userID uint,
	email string,
	level string,
) (string, error) {

	args := m.Called(
		userID,
		email,
		level,
	)

	return args.String(0),
		args.Error(1)
}

type MockGoogleTokenValidator struct {
	mock.Mock
}

func (m *MockGoogleTokenValidator) Validate(ctx context.Context, token string, audience string) (*idtoken.Payload, error) {
	args := m.Called(
		ctx,
		token,
		audience,
	)
	payload, _ := args.Get(0).(*idtoken.Payload)
	return payload, args.Error(1)
}

// MockLoginUseCase is a mock implementation of LoginUseCase
type MockLoginUseCase struct {
	mock.Mock
}

func (m *MockLoginUseCase) Execute(
	ctx context.Context,
	req *authDto.LoginRequest,
) (*authDto.LoginResponse, error) {

	args := m.Called(ctx, req)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*authDto.LoginResponse), args.Error(1)
}

type MockUserOAuthRepository struct {
	mock.Mock
}

func (m *MockUserOAuthRepository) Create(
	ctx context.Context,
	oauth *securityEntity.UserOAuth,
) error {

	args := m.Called(ctx, oauth)
	return args.Error(0)
}

func (m *MockUserOAuthRepository) GetByProviderAndUserID(
	ctx context.Context,
	provider string,
	providerUserID string,
) (*securityEntity.UserOAuth, error) {

	args := m.Called(
		ctx,
		provider,
		providerUserID,
	)

	result, _ := args.Get(0).(*securityEntity.UserOAuth)

	return result, args.Error(1)
}

func (m *MockUserOAuthRepository) GetByUserID(
	ctx context.Context,
	userID uint,
) ([]*securityEntity.UserOAuth, error) {

	args := m.Called(ctx, userID)

	result, _ := args.Get(0).([]*securityEntity.UserOAuth)

	return result, args.Error(1)
}

func (m *MockUserOAuthRepository) Delete(
	ctx context.Context,
	userID uint,
	provider string,
) error {

	args := m.Called(
		ctx,
		userID,
		provider,
	)

	return args.Error(0)
}

// MockUserOAuthService is a mock implementation of UserOAuthService
type MockUserOAuthService struct {
	mock.Mock
}

func (m *MockUserOAuthService) GetByProviderAndProviderUserID(
	ctx context.Context,
	provider string,
	providerUserID string,
) (*securityEntity.UserOAuth, error) {

	args := m.Called(
		ctx,
		provider,
		providerUserID,
	)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*securityEntity.UserOAuth), args.Error(1)
}

func (m *MockUserOAuthService) Create(
	ctx context.Context,
	oauth *securityEntity.UserOAuth,
) error {

	args := m.Called(ctx, oauth)
	return args.Error(0)
}

// MockOAuthService is a mock implementation of OAuthService
type MockOAuthService struct {
	mock.Mock
}

func (m *MockOAuthService) VerifyGoogleToken(
	ctx context.Context,
	token string,
) (*securityDto.GoogleUserInfo, error) {

	args := m.Called(ctx, token)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*securityDto.GoogleUserInfo), args.Error(1)
}

type MockGoogleLoginUseCase struct {
	mock.Mock
}

func (m *MockGoogleLoginUseCase) Execute(
	ctx context.Context,
	req *authDto.GoogleLoginRequest,
) (*authDto.GoogleLoginResponse, error) {

	args := m.Called(ctx, req)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*authDto.GoogleLoginResponse), args.Error(1)
}

type MockLogoutUseCase struct {
	mock.Mock
}

func (m *MockLogoutUseCase) Execute(
	ctx context.Context,
	req *authDto.LogoutRequest,
) error {

	args := m.Called(
		ctx,
		req,
	)

	return args.Error(0)
}

type MockPaymentNumberGenerator struct {
	mock.Mock
}

func (m *MockPaymentNumberGenerator) Generate() string {
	args := m.Called()

	if value := args.Get(0); value != nil {
		return value.(string)
	}

	return ""
}

type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) WithinTransaction(
	ctx context.Context,
	fn func(context.Context) error,
) error {

	args := m.Called(ctx)

	if fn != nil {
		return fn(ctx)
	}

	return args.Error(0)
}
