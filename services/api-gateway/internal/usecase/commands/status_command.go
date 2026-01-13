package commands

import "github.com/google/uuid"

type StatusCommand struct {
	VideoID uuid.UUID
	UserID  int64
}
