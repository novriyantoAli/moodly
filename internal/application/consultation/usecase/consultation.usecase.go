package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	authService "github.com/novriyantoAli/moodly/internal/application/authorization/service"
	"github.com/novriyantoAli/moodly/internal/application/consultation/dto"
	"github.com/novriyantoAli/moodly/internal/application/consultation/service"
	userService "github.com/novriyantoAli/moodly/internal/application/user/service"
)

type ConsultationUsecase interface {
	CreateConsultation(ctx context.Context, participantID uint, req *dto.CreateConsultationRequest) (*dto.CreateConsultationResponse, error)
	GetConsultations(userID uint) ([]dto.ConsultationResponse, error)
	GetConsultationByID(id uuid.UUID, userID uint) (*dto.ConsultationResponse, error)
	SendMessage(conversationID uuid.UUID, senderID uint, req *dto.SendMessageRequest) (*dto.MessageResponse, error)
	GetMessages(conversationID uuid.UUID, userID uint, cursor uuid.UUID, limit int) ([]dto.MessageResponse, error)
	MarkMessageRead(conversationID uuid.UUID, userID uint, req *dto.MarkMessageReadRequest) (*dto.MessageResponse, error)
	CloseConsultation(conversationID uuid.UUID, userID uint, req *dto.CloseConsultationRequest) (*dto.CloseConsultationResponse, error)
}

type consultationUsecase struct {
	consultationSvc service.ConsultationService
	userSvc         userService.UserService
	authSvc         authService.AuthorizationService
}

func NewConsultationUsecase(
	consultationSvc service.ConsultationService,
	userSvc userService.UserService,
	authSvc authService.AuthorizationService,
) ConsultationUsecase {
	return &consultationUsecase{
		consultationSvc: consultationSvc,
		userSvc:         userSvc,
		authSvc:         authSvc,
	}
}

func (u *consultationUsecase) CreateConsultation(ctx context.Context, participantID uint, req *dto.CreateConsultationRequest) (*dto.CreateConsultationResponse, error) {
	// 1. Validate if psychologist is an active user
	user, err := u.userSvc.GetUserByID(ctx, req.PsychologistID)
	if err != nil {
		return nil, errors.New("psikolog tidak ditemukan")
	}
	if !user.IsActive {
		return nil, errors.New("psikolog sedang tidak aktif")
	}

	// 2. Validate if psychologist has "psikolog" role
	isPsychologist, err := u.authSvc.IsPsychologist(ctx, req.PsychologistID)
	if err != nil {
		return nil, errors.New("gagal memvalidasi role psikolog")
	}

	if !isPsychologist {
		return nil, errors.New("pengguna yang dipilih bukan psikolog")
	}

	// 3. Create consultation using ConsultationService
	return u.consultationSvc.CreateConsultation(participantID, req)
}

func (u *consultationUsecase) GetConsultations(userID uint) ([]dto.ConsultationResponse, error) {
	return u.consultationSvc.GetConsultations(userID)
}

func (u *consultationUsecase) GetConsultationByID(id uuid.UUID, userID uint) (*dto.ConsultationResponse, error) {
	return u.consultationSvc.GetConsultationByID(id, userID)
}

func (u *consultationUsecase) SendMessage(conversationID uuid.UUID, senderID uint, req *dto.SendMessageRequest) (*dto.MessageResponse, error) {
	return u.consultationSvc.SendMessage(conversationID, senderID, req)
}

func (u *consultationUsecase) GetMessages(conversationID uuid.UUID, userID uint, cursor uuid.UUID, limit int) ([]dto.MessageResponse, error) {
	return u.consultationSvc.GetMessages(conversationID, userID, cursor, limit)
}

func (u *consultationUsecase) MarkMessageRead(conversationID uuid.UUID, userID uint, req *dto.MarkMessageReadRequest) (*dto.MessageResponse, error) {
	return u.consultationSvc.MarkMessageRead(conversationID, userID, req)
}

func (u *consultationUsecase) CloseConsultation(conversationID uuid.UUID, userID uint, req *dto.CloseConsultationRequest) (*dto.CloseConsultationResponse, error) {
	return u.consultationSvc.CloseConsultation(conversationID, userID, req)
}
