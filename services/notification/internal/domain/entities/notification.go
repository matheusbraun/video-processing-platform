package entities

import "time"

type NotificationType string
type NotificationStatus string

const (
	TypeEmail NotificationType = "EMAIL"
)

const (
	StatusPending NotificationStatus = "PENDING"
	StatusSent    NotificationStatus = "SENT"
	StatusFailed  NotificationStatus = "FAILED"
)

type Notification struct {
	ID           int64              `gorm:"primaryKey;autoIncrement"`
	UserID       int64              `gorm:"not null"`
	VideoID      *string            `gorm:"type:uuid"`
	Type         NotificationType   `gorm:"type:varchar(20);not null"`
	Status       NotificationStatus `gorm:"type:varchar(20);not null"`
	Recipient    string             `gorm:"type:text;not null"`
	Subject      *string            `gorm:"type:text"`
	ErrorMessage *string            `gorm:"type:text"`
	SentAt       *time.Time         `gorm:"type:timestamp"`
	CreatedAt    time.Time          `gorm:"autoCreateTime"`
}

func (Notification) TableName() string {
	return "notifications.notification_log"
}
