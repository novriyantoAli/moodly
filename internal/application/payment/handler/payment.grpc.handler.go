package handler

import (
	"context"

	"github.com/novriyantoAli/moodly/api/proto/payment"
	"github.com/novriyantoAli/moodly/internal/application/payment/dto"
	"github.com/novriyantoAli/moodly/internal/application/payment/entity"
	"github.com/novriyantoAli/moodly/internal/application/payment/service"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentGrpcHandler struct {
	payment.UnimplementedPaymentServiceServer
	paymentService service.PaymentService
	logger         *zap.Logger
}

func NewPaymentGrpcHandler(paymentService service.PaymentService, logger *zap.Logger) *PaymentGrpcHandler {
	return &PaymentGrpcHandler{
		paymentService: paymentService,
		logger:         logger,
	}
}

func (h *PaymentGrpcHandler) CreatePayment(
	ctx context.Context,
	req *payment.CreatePaymentRequest,
) (*payment.PaymentResponse, error) {
	createReq := &dto.CreatePaymentRequest{
		Method:     req.Method,
		BillNumber: req.BillNumber,
	}

	paymentResponse, err := h.paymentService.CreatePayment(ctx, createReq)
	if err != nil {
		h.logger.Error("Failed to create payment via gRPC", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create payment: %v", err)
	}

	return &payment.PaymentResponse{
		Data:    h.toProtoPayment(paymentResponse),
		Message: "Payment created successfully",
	}, nil
}

func (h *PaymentGrpcHandler) GetPayment(
	ctx context.Context,
	req *payment.GetPaymentRequest,
) (*payment.PaymentResponse, error) {
	paymentResponse, err := h.paymentService.GetPaymentByID(ctx, uint(req.Id))
	if err != nil {
		h.logger.Error("Failed to get payment via gRPC", zap.Uint32("id", req.Id), zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "payment not found: %v", err)
	}

	return &payment.PaymentResponse{
		Data:    h.toProtoPayment(paymentResponse),
		Message: "Payment retrieved successfully",
	}, nil
}

func (h *PaymentGrpcHandler) ListPayments(
	ctx context.Context,
	req *payment.ListPaymentsRequest,
) (*payment.ListPaymentsResponse, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	filter := &dto.PaymentFilter{
		Page:          page,
		PageSize:      pageSize,
		Currency:      req.Currency,
		BillID:        req.BillId,
		PaymentNumber: req.PaymentNumber,
		Method:        req.Method,
	}

	// Add status filter if provided
	if req.Status != payment.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED {
		filter.Status = h.protoStatusToString(req.Status)
	}

	listResponse, err := h.paymentService.GetPayments(ctx, filter)
	if err != nil {
		h.logger.Error("Failed to list payments via gRPC", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list payments: %v", err)
	}

	protoPayments := make([]*payment.Payment, len(listResponse.Data))
	for i, p := range listResponse.Data {
		protoPayments[i] = h.toProtoPayment(&p)
	}

	return &payment.ListPaymentsResponse{
		Payments:   protoPayments,
		TotalCount: listResponse.TotalCount,
		Page:       int32(listResponse.Page),
		PageSize:   int32(listResponse.PageSize),
	}, nil
}

func (h *PaymentGrpcHandler) UpdatePayment(
	ctx context.Context,
	req *payment.UpdatePaymentRequest,
) (*payment.PaymentResponse, error) {
	updateReq := &dto.UpdatePaymentRequest{
		Description: req.Description,
	}

	// Add status if provided
	if req.Status != payment.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED {
		updateReq.Status = h.protoStatusToString(req.Status)
	}

	paymentResponse, err := h.paymentService.UpdatePayment(ctx, uint(req.Id), updateReq)
	if err != nil {
		h.logger.Error("Failed to update payment via gRPC", zap.Uint32("id", req.Id), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update payment: %v", err)
	}

	return &payment.PaymentResponse{
		Data:    h.toProtoPayment(paymentResponse),
		Message: "Payment updated successfully",
	}, nil
}

func (h *PaymentGrpcHandler) toProtoPayment(p *dto.PaymentResponse) *payment.Payment {
	protoPayment := &payment.Payment{
		Id:               uint32(p.ID),
		Amount:           float64(p.Amount),
		Currency:         p.Currency,
		Description:      p.Description,
		Status:           h.stringStatusToProto(p.Status),
		CreatedBy:        uint32(p.CreatedBy),
		CreatedAt:        timestamppb.New(p.CreatedAt),
		UpdatedAt:        timestamppb.New(p.UpdatedAt),
		PaymentNumber:    p.PaymentNumber,
		BillId:           p.BillID,
		Method:           p.Method,
		GatewayReference: p.GatewayReference,
	}

	if p.PaidAt != nil {
		protoPayment.PaidAt = timestamppb.New(*p.PaidAt)
	}

	return protoPayment
}

func (h *PaymentGrpcHandler) stringStatusToProto(status string) payment.PaymentStatus {
	switch status {
	case entity.PaymentStatusPending.String():
		return payment.PaymentStatus_PAYMENT_STATUS_PENDING
	case entity.PaymentStatusCompleted.String():
		return payment.PaymentStatus_PAYMENT_STATUS_COMPLETED
	case entity.PaymentStatusFailed.String():
		return payment.PaymentStatus_PAYMENT_STATUS_FAILED
	case entity.PaymentStatusCanceled.String():
		return payment.PaymentStatus_PAYMENT_STATUS_CANCELED
	default:
		return payment.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}

func (h *PaymentGrpcHandler) protoStatusToString(status payment.PaymentStatus) string {
	switch status {
	case payment.PaymentStatus_PAYMENT_STATUS_PENDING:
		return entity.PaymentStatusPending.String()
	case payment.PaymentStatus_PAYMENT_STATUS_COMPLETED:
		return entity.PaymentStatusCompleted.String()
	case payment.PaymentStatus_PAYMENT_STATUS_FAILED:
		return entity.PaymentStatusFailed.String()
	case payment.PaymentStatus_PAYMENT_STATUS_CANCELED:
		return entity.PaymentStatusCanceled.String()
	default:
		return entity.PaymentStatusPending.String()
	}
}
