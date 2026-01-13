package entities

import (
	"time"

	"github.com/google/uuid"
)

type VideoStatus string

const (
	StatusPending    VideoStatus = "PENDING"
	StatusProcessing VideoStatus = "PROCESSING"
	StatusCompleted  VideoStatus = "COMPLETED"
	StatusFailed     VideoStatus = "FAILED"
)

type Video struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       int64       `gorm:"not null;index:idx_user_status"`
	Filename     string      `gorm:"type:varchar(255);not null"`
	OriginalPath string      `gorm:"type:text;not null"`
	Status       VideoStatus `gorm:"type:varchar(20);not null;index:idx_user_status"`
	FPS          int         `gorm:"default:1"`
	FrameCount   *int        `gorm:"type:int"`
	ZipPath      *string     `gorm:"type:text"`
	ErrorMessage *string     `gorm:"type:text"`
	CreatedAt    time.Time   `gorm:"autoCreateTime;index:idx_created_at"`
	StartedAt    *time.Time  `gorm:"type:timestamp"`
	CompletedAt  *time.Time  `gorm:"type:timestamp"`
	ExpiresAt    time.Time   `gorm:"type:timestamp;index:idx_expires_at"`
}

func (Video) TableName() string {
	return "videos.videos"
}
