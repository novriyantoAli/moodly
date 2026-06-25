package entity

import (
	"time"

	"github.com/google/uuid"
	userEntity "github.com/novriyantoAli/moodly/internal/application/user/entity"
)

type ConsultationStatus string

const (
	StatusWaiting   ConsultationStatus = "WAITING"
	StatusActive    ConsultationStatus = "ACTIVE"
	StatusClosed    ConsultationStatus = "CLOSED"
	StatusCancelled ConsultationStatus = "CANCELLED"
)

type MessageType string

const (
	MessageTypeText   MessageType = "TEXT"
	MessageTypeImage  MessageType = "IMAGE"
	MessageTypeFile   MessageType = "FILE"
	MessageTypeSystem MessageType = "SYSTEM"
)

type Conversation struct {
	ID             uuid.UUID          `gorm:"primaryKey;type:uuid"`
	ParticipantID  uint               `gorm:"not null"`
	Participant    userEntity.User    `gorm:"foreignKey:ParticipantID"`
	PsychologistID uint               `gorm:"not null"`
	Psychologist   userEntity.User    `gorm:"foreignKey:PsychologistID"`
	Status         ConsultationStatus `gorm:"type:varchar(20);default:'WAITING'"`
	StartedAt      *time.Time
	ClosedAt       *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (c Conversation) TableName() string {
	return "conversations"
}

type Message struct {
	ID             uuid.UUID       `gorm:"primaryKey;type:uuid"`
	ConversationID uuid.UUID       `gorm:"type:uuid;not null"`
	Conversation   Conversation    `gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE;"`
	SenderID       uint            `gorm:"not null"`
	Sender         userEntity.User `gorm:"foreignKey:SenderID"`
	MessageType    MessageType     `gorm:"type:varchar(20);default:'TEXT'"`
	Message        string          `gorm:"type:text"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (m Message) TableName() string {
	return "messages"
}

type MessageAttachment struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid"`
	MessageID   uuid.UUID `gorm:"type:uuid;not null"`
	Message     Message   `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE;"`
	FileName    string    `gorm:"type:text"`
	FileUrl     string    `gorm:"type:text"`
	ContentType string    `gorm:"type:text"`
	Size        int
	CreatedAt   time.Time
}

func (ma MessageAttachment) TableName() string {
	return "message_attachments"
}

type MessageRead struct {
	ID        uuid.UUID       `gorm:"primaryKey;type:uuid"`
	MessageID uuid.UUID       `gorm:"type:uuid;not null"`
	Message   Message         `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE;"`
	UserID    uint            `gorm:"not null"`
	User      userEntity.User `gorm:"foreignKey:UserID"`
	CreatedAt time.Time
}

func (mr MessageRead) TableName() string {
	return "message_reads"
}
