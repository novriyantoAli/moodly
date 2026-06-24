package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/novriyantoAli/moodly/internal/application/consultation/entity"
)

type CreateConsultationRequest struct {
	PsychologistID uint `json:"psychologist_id" binding:"required"`
}

type CreateConsultationResponse struct {
	ConversationID uuid.UUID                 `json:"conversation_id"`
	Status         entity.ConsultationStatus `json:"status"`
}

type ConsultationResponse struct {
	ConversationID uuid.UUID                 `json:"conversation_id"`
	PsychologistID uint                      `json:"psychologist_id"`
	Status         entity.ConsultationStatus `json:"status"`
	StartedAt      *time.Time                `json:"started_at"`
	ClosedAt       *time.Time                `json:"closed_at"`
	CreatedAt      time.Time                 `json:"created_at"`
}

type SendMessageRequest struct {
	Message     string             `json:"message" binding:"required"`
	MessageType entity.MessageType `json:"message_type" binding:"required,oneof=TEXT IMAGE FILE SYSTEM"`
}

type MessageResponse struct {
	MessageID      uuid.UUID          `json:"message_id"`
	ConversationID uuid.UUID          `json:"conversation_id"`
	SenderID       uint               `json:"sender_id"`
	MessageType    entity.MessageType `json:"message_type"`
	Message        string             `json:"message"`
	CreatedAt      time.Time          `json:"created_at"`
}

type MarkMessageReadRequest struct {
	MessageID uuid.UUID `json:"message_id" binding:"required"`
}

type CloseConsultationRequest struct {
	Status entity.ConsultationStatus `json:"status" binding:"required,eq=CLOSED"`
}

type CloseConsultationResponse struct {
	ConversationID uuid.UUID                 `json:"conversation_id"`
	Status         entity.ConsultationStatus `json:"status"`
}
