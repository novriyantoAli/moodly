package handler

import (
	"context"

	"github.com/novriyantoAli/moodly/api/proto/user"
	"github.com/novriyantoAli/moodly/internal/application/user/dto"
	"github.com/novriyantoAli/moodly/internal/application/user/service"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserGrpcHandler struct {
	user.UnimplementedUserServiceServer
	userService service.UserService
	logger      *zap.Logger
}

func NewUserGrpcHandler(userService service.UserService, logger *zap.Logger) *UserGrpcHandler {
	return &UserGrpcHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserGrpcHandler) CreateUser(
	ctx context.Context,
	req *user.CreateUserRequest,
) (*user.UserResponse, error) {
	createReq := &dto.CreateUserRequest{
		Email:    req.Email,
		FullName: req.FullName,
	}

	userResponse, err := h.userService.CreateUser(ctx, createReq)
	if err != nil {
		h.logger.Error("Failed to create user via gRPC", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &user.UserResponse{
		Data:    h.toProtoUser(userResponse),
		Message: "User created successfully",
	}, nil
}

func (h *UserGrpcHandler) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.UserResponse, error) {
	userResponse, err := h.userService.GetUserByID(ctx, uint(req.Id))
	if err != nil {
		h.logger.Error("Failed to get user via gRPC", zap.Uint32("id", req.Id), zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &user.UserResponse{
		Data:    h.toProtoUser(userResponse),
		Message: "User retrieved successfully",
	}, nil
}

func (h *UserGrpcHandler) GetUserByEmail(ctx context.Context, req *user.GetUserByEmailRequest) (*user.UserResponse, error) {
	userResponse, err := h.userService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		h.logger.Error("Failed to get user by email via gRPC", zap.String("email", req.Email), zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &user.UserResponse{
		Data:    h.toProtoUser(userResponse),
		Message: "User retrieved successfully",
	}, nil
}

func (h *UserGrpcHandler) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	filter := &dto.UserFilter{
		Page:     page,
		PageSize: pageSize,
	}

	listResponse, err := h.userService.GetUsers(ctx, filter)
	if err != nil {
		h.logger.Error("Failed to list users via gRPC", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	protoUsers := make([]*user.User, len(listResponse.Data))
	for i, u := range listResponse.Data {
		protoUsers[i] = h.toProtoUser(&u)
	}

	return &user.ListUsersResponse{
		Users:      protoUsers,
		TotalCount: listResponse.TotalCount,
		Page:       int32(listResponse.Page),
		PageSize:   int32(listResponse.PageSize),
	}, nil
}

func (h *UserGrpcHandler) UpdateUser(
	ctx context.Context,
	req *user.UpdateUserRequest,
) (*user.UserResponse, error) {
	updateReq := &dto.UpdateUserRequest{
		FullName: req.FullName,
		Level:    req.Level,
		IsActive: req.IsActive,
	}

	userResponse, err := h.userService.UpdateUser(ctx, uint(req.Id), updateReq)
	if err != nil {
		h.logger.Error("Failed to update user via gRPC", zap.Uint32("id", req.Id), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &user.UserResponse{
		Data:    h.toProtoUser(userResponse),
		Message: "User updated successfully",
	}, nil
}

func (h *UserGrpcHandler) DeleteUser(
	ctx context.Context,
	req *user.DeleteUserRequest,
) (*user.UserResponse, error) {
	err := h.userService.DeleteUser(ctx, uint(req.Id))
	if err != nil {
		h.logger.Error("Failed to delete user via gRPC", zap.Uint32("id", req.Id), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &user.UserResponse{
		Message: "User deleted successfully",
	}, nil
}

func (h *UserGrpcHandler) toProtoUser(u *dto.UserResponse) *user.User {
	return &user.User{
		Id:        uint32(u.ID),
		Email:     u.Email,
		FullName:  u.FullName,
		Level:     u.Level,
		IsActive:  u.IsActive,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}
