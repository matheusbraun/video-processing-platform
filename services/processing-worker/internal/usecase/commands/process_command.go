package commands

import "github.com/google/uuid"

type ProcessCommand struct {
	VideoID   uuid.UUID
	UserID    int64
	UserEmail string
	S3Key     string
	Filename  string
}
