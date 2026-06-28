package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/novriyantoAli/moodly/internal/application/consultation/dto"
	"github.com/novriyantoAli/moodly/internal/application/consultation/entity"
	"github.com/novriyantoAli/moodly/internal/application/consultation/repository"
)

type ConsultationService interface {
	ValidateConsultation(ctx context.Context, id uuid.UUID, userID uint) (bool, error)
	CreateConsultation(ctx context.Context, participantID uint, req *dto.CreateConsultationRequest) (*dto.CreateConsultationResponse, error)
	GetConsultations(ctx context.Context, userID uint, limit int, page int, status string, search string) ([]dto.ConsultationResponse, int64, error)
	GetConsultationByID(ctx context.Context, id uuid.UUID, userID uint) (*dto.ConsultationResponse, error)
	SendMessage(ctx context.Context, conversationID uuid.UUID, senderID uint, req *dto.SendMessageRequest) (*dto.MessageResponse, error)
	GetMessages(ctx context.Context, conversationID uuid.UUID, userID uint, cursor uuid.UUID, limit int) ([]dto.MessageResponse, error)
	MarkMessageRead(ctx context.Context, conversationID uuid.UUID, userID uint, req *dto.MarkMessageReadRequest) (*dto.MessageResponse, error)
	ApproveConsultation(ctx context.Context, conversationID uuid.UUID, psychologistID uint) error
	CloseConsultation(ctx context.Context, conversationID uuid.UUID, userID uint) (*dto.CloseConsultationResponse, error)
}

type consultationService struct {
	repo repository.ConsultationRepository
}

func NewConsultationService(repo repository.ConsultationRepository) ConsultationService {
	return &consultationService{repo: repo}
}

func (s *consultationService) ValidateConsultation(ctx context.Context, id uuid.UUID, userID uint) (bool, error) {
	conversation, err := s.repo.GetConversationByID(ctx, id)
	if err != nil {
		return false, err
	}
	if conversation.Status != entity.StatusActive {
		return false, errors.New("sesi konsultasi ini sudah berakhir")
	}
	if conversation.ParticipantID != userID && conversation.PsychologistID != userID {
		return false, errors.New("anda tidak memiliki akses ke sesi konsultasi ini")
	}
	return true, nil
}

func (s *consultationService) CreateConsultation(ctx context.Context, participantID uint, req *dto.CreateConsultationRequest) (*dto.CreateConsultationResponse, error) {
	exists, err := s.repo.ExistsActiveConsultation(ctx, participantID)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New("you have an active consultation")
	}

	conv := &entity.Conversation{
		ID:             uuid.New(),
		ParticipantID:  participantID,
		PsychologistID: req.PsychologistID,
		Status:         entity.StatusWaiting,
	}

	if err := s.repo.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}

	return &dto.CreateConsultationResponse{
		ConversationID: conv.ID,
		Status:         conv.Status,
	}, nil
}

func (s *consultationService) GetConsultations(ctx context.Context, userID uint, limit int, page int, status string, search string) ([]dto.ConsultationResponse, int64, error) {
	offset := (page - 1) * limit
	convs, totalCount, err := s.repo.GetConversations(ctx, userID, limit, offset, status, search)
	if err != nil {
		return nil, 0, err
	}

	var responses []dto.ConsultationResponse
	for _, c := range convs {
		responses = append(responses, dto.ConsultationResponse{
			ConversationID: c.ID,
			PsychologistID: c.PsychologistID,
			Status:         c.Status,
			StartedAt:      c.StartedAt,
			ClosedAt:       c.ClosedAt,
			CreatedAt:      c.CreatedAt,
			Participant: dto.UserDetailResponse{
				ID:    c.Participant.ID,
				Name:  c.Participant.FullName,
				Email: c.Participant.Email,
			},
			Psychologist: dto.UserDetailResponse{
				ID:    c.Psychologist.ID,
				Name:  c.Psychologist.FullName,
				Email: c.Psychologist.Email,
			},
		})
	}
	return responses, totalCount, nil
}

func (s *consultationService) GetConsultationByID(ctx context.Context, id uuid.UUID, userID uint) (*dto.ConsultationResponse, error) {
	c, err := s.repo.GetConversationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.ParticipantID != userID && c.PsychologistID != userID {
		return nil, errors.New("unauthorized access to conversation")
	}

	return &dto.ConsultationResponse{
		ConversationID: c.ID,
		PsychologistID: c.PsychologistID,
		Status:         c.Status,
		StartedAt:      c.StartedAt,
		ClosedAt:       c.ClosedAt,
		CreatedAt:      c.CreatedAt,
	}, nil
}



func (s *consultationService) SendMessage(ctx context.Context, conversationID uuid.UUID, senderID uint, req *dto.SendMessageRequest) (*dto.MessageResponse, error) {
	c, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if c.ParticipantID != senderID && c.PsychologistID != senderID {
		return nil, errors.New("unauthorized to send message in this conversation")
	}

	msg := &entity.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       senderID,
		MessageType:    req.MessageType,
		Message:        req.Message,
	}

	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}

	return &dto.MessageResponse{
		MessageID:      msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		MessageType:    msg.MessageType,
		Message:        msg.Message,
		CreatedAt:      msg.CreatedAt,
	}, nil
}

func (s *consultationService) GetMessages(ctx context.Context, conversationID uuid.UUID, userID uint, cursor uuid.UUID, limit int) ([]dto.MessageResponse, error) {
	c, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if c.ParticipantID != userID && c.PsychologistID != userID {
		return nil, errors.New("unauthorized to view messages in this conversation")
	}

	msgs, err := s.repo.GetMessages(ctx, conversationID, cursor, limit)
	if err != nil {
		return nil, err
	}

	var responses []dto.MessageResponse
	for _, m := range msgs {
		responses = append(responses, dto.MessageResponse{
			MessageID:      m.ID,
			ConversationID: m.ConversationID,
			SenderID:       m.SenderID,
			MessageType:    m.MessageType,
			Message:        m.Message,
			CreatedAt:      m.CreatedAt,
		})
	}
	return responses, nil
}

func (s *consultationService) MarkMessageRead(ctx context.Context, conversationID uuid.UUID, userID uint, req *dto.MarkMessageReadRequest) (*dto.MessageResponse, error) {
	msg, err := s.repo.GetMessageByID(ctx, req.MessageID)
	if err != nil {
		return nil, err
	}

	if msg.ConversationID != conversationID {
		return nil, errors.New("message does not belong to conversation")
	}

	read := &entity.MessageRead{
		MessageID: req.MessageID,
		UserID:    userID,
	}

	if err := s.repo.MarkMessageAsRead(ctx, read); err != nil {
		return nil, err
	}

	return &dto.MessageResponse{
		MessageID:      msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		MessageType:    msg.MessageType,
		Message:        msg.Message,
		CreatedAt:      msg.CreatedAt,
	}, nil
}

func (s *consultationService) ApproveConsultation(ctx context.Context, conversationID uuid.UUID, psychologistID uint) error {
	c, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return err
	}

	if c.PsychologistID != psychologistID {
		return errors.New("unauthorized to approve this conversation")
	}

	if c.Status != entity.StatusWaiting {
		return errors.New("conversation is already active or closed")
	}

	if err := s.repo.UpdateConversationStatus(ctx, conversationID, entity.StatusActive); err != nil {
		return err
	}

	return nil
}

func (s *consultationService) CloseConsultation(ctx context.Context, conversationID uuid.UUID, userID uint) (*dto.CloseConsultationResponse, error) {
	if err := s.repo.UpdateConversationStatus(ctx, conversationID, entity.StatusClosed); err != nil {
		return nil, err
	}

	return &dto.CloseConsultationResponse{
		ConversationID: conversationID,
		Status:         entity.StatusClosed,
	}, nil
}
