package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/novriyantoAli/moodly/internal/application/consultation/entity"
	"gorm.io/gorm"
)

type ConsultationRepository interface {
	CreateConversation(ctx context.Context, conversation *entity.Conversation) error
	GetConversations(ctx context.Context, userID uint) ([]entity.Conversation, error)
	GetConversationByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error)
	UpdateConversationStatus(ctx context.Context, id uuid.UUID, status entity.ConsultationStatus) error
	UpdateConversation(ctx context.Context, conversation *entity.Conversation) error
	ExistsActiveConsultation(ctx context.Context, participantID uint) (bool, error)

	CreateMessage(ctx context.Context, message *entity.Message) error
	GetMessageByID(ctx context.Context, id uuid.UUID) (*entity.Message, error)
	GetMessages(ctx context.Context, conversationID uuid.UUID, cursor uuid.UUID, limit int) ([]entity.Message, error)
	MarkMessageAsRead(ctx context.Context, read *entity.MessageRead) error
}

type consultationRepository struct {
	db *gorm.DB
}

func NewConsultationRepository(db *gorm.DB) ConsultationRepository {
	return &consultationRepository{db: db}
}

func (r *consultationRepository) CreateConversation(ctx context.Context, conversation *entity.Conversation) error {
	
	return r.db.WithContext(ctx).Create(conversation).Error
}

func (r *consultationRepository) GetConversations(ctx context.Context, userID uint) ([]entity.Conversation, error) {
	var conversations []entity.Conversation
	err := r.db.WithContext(ctx).Where("participant_id = ? OR psychologist_id = ?", userID, userID).
		Order("updated_at DESC").
		Find(&conversations).Error
	return conversations, err
}

func (r *consultationRepository) GetConversationByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	var conversation entity.Conversation
	err := r.db.WithContext(ctx).First(&conversation, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *consultationRepository) UpdateConversationStatus(ctx context.Context, id uuid.UUID, status entity.ConsultationStatus) error {
	return r.db.WithContext(ctx).Model(&entity.Conversation{}).Where("id = ?", id).Update("status", status).Error
}

func (r *consultationRepository) UpdateConversation(ctx context.Context, conversation *entity.Conversation) error {
	return r.db.WithContext(ctx).Save(conversation).Error
}

// func (r *consultationRepository) ExistsActiveConsultation(ctx context.Context, participantID uint) (bool, error) {
// 	var count int64
// 	err := r.db.WithContext(ctx).Model(&entity.Conversation{}).
// 		Where("participant_id = ? AND (status = ? OR status = ?)", participantID, entity.StatusWaiting, entity.StatusActive).
// 		Count(&count).Error
// 	if err != nil {
// 		return false, err
// 	}
// 	return count > 0, nil
// }

func (r *consultationRepository) ExistsActiveConsultation(
    ctx context.Context,
    participantID uint,
) (bool, error) {
    var exists bool

    err := r.db.WithContext(ctx).
        Raw(`
            SELECT EXISTS (
                SELECT 1
                FROM conversations
                WHERE participant_id = ?
                  AND status IN (?, ?)
            )
        `,
            participantID,
            entity.StatusWaiting,
            entity.StatusActive,
        ).
        Scan(&exists).
        Error

    return exists, err
}

func (r *consultationRepository) CreateMessage(ctx context.Context, message *entity.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *consultationRepository) GetMessageByID(ctx context.Context, id uuid.UUID) (*entity.Message, error) {
	var message entity.Message
	err := r.db.WithContext(ctx).First(&message, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *consultationRepository) GetMessages(ctx context.Context, conversationID uuid.UUID, cursor uuid.UUID, limit int) ([]entity.Message, error) {
	var messages []entity.Message
	query := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Order("created_at DESC")

	if cursor != uuid.Nil {
		var cursorMessage entity.Message
		if err := r.db.WithContext(ctx).First(&cursorMessage, "id = ?", cursor).Error; err == nil {
			query = query.Where("created_at < ?", cursorMessage.CreatedAt)
		}
	}

	err := query.Limit(limit).Find(&messages).Error
	return messages, err
}

func (r *consultationRepository) MarkMessageAsRead(ctx context.Context, read *entity.MessageRead) error {
	return r.db.WithContext(ctx).Create(read).Error
}
