package repository

import (
	"github.com/google/uuid"
	"github.com/novriyantoAli/moodly/internal/application/consultation/entity"
	"gorm.io/gorm"
)

type ConsultationRepository interface {
	CreateConversation(conversation *entity.Conversation) error
	GetConversations(userID uint) ([]entity.Conversation, error)
	GetConversationByID(id uuid.UUID) (*entity.Conversation, error)
	UpdateConversationStatus(id uuid.UUID, status entity.ConsultationStatus) error
	
	CreateMessage(message *entity.Message) error
	GetMessageByID(id uuid.UUID) (*entity.Message, error)
	GetMessages(conversationID uuid.UUID, cursor uuid.UUID, limit int) ([]entity.Message, error)
	MarkMessageAsRead(read *entity.MessageRead) error
}

type consultationRepository struct {
	db *gorm.DB
}

func NewConsultationRepository(db *gorm.DB) ConsultationRepository {
	return &consultationRepository{db: db}
}

func (r *consultationRepository) CreateConversation(conversation *entity.Conversation) error {
	return r.db.Create(conversation).Error
}

func (r *consultationRepository) GetConversations(userID uint) ([]entity.Conversation, error) {
	var conversations []entity.Conversation
	err := r.db.Where("participant_id = ? OR psychologist_id = ?", userID, userID).
		Order("updated_at DESC").
		Find(&conversations).Error
	return conversations, err
}

func (r *consultationRepository) GetConversationByID(id uuid.UUID) (*entity.Conversation, error) {
	var conversation entity.Conversation
	err := r.db.First(&conversation, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *consultationRepository) UpdateConversationStatus(id uuid.UUID, status entity.ConsultationStatus) error {
	return r.db.Model(&entity.Conversation{}).Where("id = ?", id).Update("status", status).Error
}

func (r *consultationRepository) CreateMessage(message *entity.Message) error {
	return r.db.Create(message).Error
}

func (r *consultationRepository) GetMessageByID(id uuid.UUID) (*entity.Message, error) {
	var message entity.Message
	err := r.db.First(&message, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *consultationRepository) GetMessages(conversationID uuid.UUID, cursor uuid.UUID, limit int) ([]entity.Message, error) {
	var messages []entity.Message
	query := r.db.Where("conversation_id = ?", conversationID).Order("created_at DESC")
	
	if cursor != uuid.Nil {
		var cursorMessage entity.Message
		if err := r.db.First(&cursorMessage, "id = ?", cursor).Error; err == nil {
			query = query.Where("created_at < ?", cursorMessage.CreatedAt)
		}
	}
	
	err := query.Limit(limit).Find(&messages).Error
	return messages, err
}

func (r *consultationRepository) MarkMessageAsRead(read *entity.MessageRead) error {
	return r.db.Create(read).Error
}
