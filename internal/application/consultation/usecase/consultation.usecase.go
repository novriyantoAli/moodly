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
	GetConsultations(ctx context.Context, userID uint, limit int, page int, status string, search string) ([]dto.ConsultationResponse, int64, error)
	GetConsultationByID(ctx context.Context, id uuid.UUID, userID uint) (*dto.ConsultationResponse, error)
	SendMessage(ctx context.Context, conversationID uuid.UUID, senderID uint, req *dto.SendMessageRequest) (*dto.MessageResponse, error)
	GetMessages(ctx context.Context, conversationID uuid.UUID, userID uint, cursor uuid.UUID, limit int) ([]dto.MessageResponse, error)
	MarkMessageRead(ctx context.Context, conversationID uuid.UUID, userID uint, req *dto.MarkMessageReadRequest) (*dto.MessageResponse, error)
	ApproveConsultation(ctx context.Context, conversationID uuid.UUID, psychologistID uint) error
	CloseConsultation(ctx context.Context, conversationID uuid.UUID, userID uint) (*dto.CloseConsultationResponse, error)
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
	user, err := u.userSvc.ValidateUser(ctx, req.PsychologistID)
	if err != nil {
		return nil, err
	}

	isPsychologist, err := u.authSvc.IsPsychologist(ctx, user.ID)
	if err != nil {
		return nil, errors.New("gagal memvalidasi role psikolog")
	}

	if !isPsychologist {
		return nil, errors.New("pengguna yang dipilih bukan psikolog")
	}

	// 3. Create consultation using ConsultationService
	return u.consultationSvc.CreateConsultation(ctx, participantID, req)
}

func (u *consultationUsecase) GetConsultations(ctx context.Context, userID uint, limit int, page int, status string, search string) ([]dto.ConsultationResponse, int64, error) {
	return u.consultationSvc.GetConsultations(ctx, userID, limit, page, status, search)
}

func (u *consultationUsecase) GetConsultationByID(ctx context.Context, id uuid.UUID, userID uint) (*dto.ConsultationResponse, error) {
	return u.consultationSvc.GetConsultationByID(ctx, id, userID)
}

func (u *consultationUsecase) SendMessage(ctx context.Context, conversationID uuid.UUID, senderID uint, req *dto.SendMessageRequest) (*dto.MessageResponse, error) {
	user, err := u.userSvc.ValidateUser(ctx, senderID)
	if err != nil {
		return nil, err
	}

	valid, err := u.consultationSvc.ValidateConsultation(ctx, conversationID, user.ID)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, errors.New("anda tidak memiliki akses ke sesi konsultasi ini")
	}

	return u.consultationSvc.SendMessage(ctx, conversationID, user.ID, req)
}

func (u *consultationUsecase) GetMessages(ctx context.Context, conversationID uuid.UUID, userID uint, cursor uuid.UUID, limit int) ([]dto.MessageResponse, error) {
	return u.consultationSvc.GetMessages(ctx, conversationID, userID, cursor, limit)
}

func (u *consultationUsecase) MarkMessageRead(ctx context.Context, conversationID uuid.UUID, userID uint, req *dto.MarkMessageReadRequest) (*dto.MessageResponse, error) {
	return u.consultationSvc.MarkMessageRead(ctx, conversationID, userID, req)
}

func (u *consultationUsecase) CloseConsultation(ctx context.Context, conversationID uuid.UUID, userID uint) (*dto.CloseConsultationResponse, error) {
	isVal, err := u.consultationSvc.ValidateConsultation(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !isVal {
		return nil, errors.New("anda tidak memiliki sesi konsultasi yang aktif")
	}

	return u.consultationSvc.CloseConsultation(ctx, conversationID, userID)
}

func (u *consultationUsecase) ApproveConsultation(ctx context.Context, conversationID uuid.UUID, psychologistID uint) error {
	user, err := u.userSvc.ValidateUser(ctx, psychologistID)
	if err != nil {
		return errors.New("psikolog tidak ditemukan")
	}

	isPsychologist, err := u.authSvc.IsPsychologist(ctx, user.ID)
	if err != nil {
		return errors.New("gagal memvalidasi role psikolog")
	}
	if !isPsychologist {
		return errors.New("pengguna yang menyetujui sesi bukan psikolog")
	}

	return u.consultationSvc.ApproveConsultation(ctx, conversationID, psychologistID)
}
