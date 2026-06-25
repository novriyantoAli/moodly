package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/novriyantoAli/moodly/internal/application/consultation/dto"
	"github.com/novriyantoAli/moodly/internal/application/consultation/entity"
	"github.com/novriyantoAli/moodly/internal/application/consultation/repository"
)

type ConsultationService interface {
	CreateConsultation(participantID uint, req *dto.CreateConsultationRequest) (*dto.CreateConsultationResponse, error)
	GetConsultations(userID uint) ([]dto.ConsultationResponse, error)
	GetConsultationByID(id uuid.UUID, userID uint) (*dto.ConsultationResponse, error)
	SendMessage(conversationID uuid.UUID, senderID uint, req *dto.SendMessageRequest) (*dto.MessageResponse, error)
	GetMessages(conversationID uuid.UUID, userID uint, cursor uuid.UUID, limit int) ([]dto.MessageResponse, error)
	MarkMessageRead(conversationID uuid.UUID, userID uint, req *dto.MarkMessageReadRequest) (*dto.MessageResponse, error)
	CloseConsultation(conversationID uuid.UUID, userID uint, req *dto.CloseConsultationRequest) (*dto.CloseConsultationResponse, error)
}

type consultationService struct {
	repo repository.ConsultationRepository
}

func NewConsultationService(repo repository.ConsultationRepository) ConsultationService {
	return &consultationService{repo: repo}
}

func (s *consultationService) CreateConsultation(participantID uint, req *dto.CreateConsultationRequest) (*dto.CreateConsultationResponse, error) {
	conv := &entity.Conversation{
		ParticipantID:  participantID,
		PsychologistID: req.PsychologistID,
		Status:         entity.StatusWaiting,
	}

	if err := s.repo.CreateConversation(conv); err != nil {
		return nil, err
	}

	return &dto.CreateConsultationResponse{
		ConversationID: conv.ID,
		Status:         conv.Status,
	}, nil
}

func (s *consultationService) GetConsultations(userID uint) ([]dto.ConsultationResponse, error) {
	convs, err := s.repo.GetConversations(userID)
	if err != nil {
		return nil, err
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
		})
	}
	return responses, nil
}

func (s *consultationService) GetConsultationByID(id uuid.UUID, userID uint) (*dto.ConsultationResponse, error) {
	c, err := s.repo.GetConversationByID(id)
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

func (s *consultationService) SendMessage(conversationID uuid.UUID, senderID uint, req *dto.SendMessageRequest) (*dto.MessageResponse, error) {
	c, err := s.repo.GetConversationByID(conversationID)
	if err != nil {
		return nil, err
	}
	if c.ParticipantID != senderID && c.PsychologistID != senderID {
		return nil, errors.New("unauthorized to send message in this conversation")
	}

	msg := &entity.Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		MessageType:    req.MessageType,
		Message:        req.Message,
	}

	if err := s.repo.CreateMessage(msg); err != nil {
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

func (s *consultationService) GetMessages(conversationID uuid.UUID, userID uint, cursor uuid.UUID, limit int) ([]dto.MessageResponse, error) {
	c, err := s.repo.GetConversationByID(conversationID)
	if err != nil {
		return nil, err
	}
	if c.ParticipantID != userID && c.PsychologistID != userID {
		return nil, errors.New("unauthorized to view messages in this conversation")
	}

	msgs, err := s.repo.GetMessages(conversationID, cursor, limit)
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

func (s *consultationService) MarkMessageRead(conversationID uuid.UUID, userID uint, req *dto.MarkMessageReadRequest) (*dto.MessageResponse, error) {
	msg, err := s.repo.GetMessageByID(req.MessageID)
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

	if err := s.repo.MarkMessageAsRead(read); err != nil {
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

func (s *consultationService) CloseConsultation(conversationID uuid.UUID, userID uint, req *dto.CloseConsultationRequest) (*dto.CloseConsultationResponse, error) {
	c, err := s.repo.GetConversationByID(conversationID)
	if err != nil {
		return nil, err
	}

	if c.ParticipantID != userID && c.PsychologistID != userID {
		return nil, errors.New("unauthorized to close this conversation")
	}

	if err := s.repo.UpdateConversationStatus(conversationID, entity.StatusClosed); err != nil {
		return nil, err
	}

	return &dto.CloseConsultationResponse{
		ConversationID: conversationID,
		Status:         entity.StatusClosed,
	}, nil
}
