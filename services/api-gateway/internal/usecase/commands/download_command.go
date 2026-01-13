package commands

import "github.com/google/uuid"

type DownloadCommand struct {
	VideoID uuid.UUID
	UserID  int64
}
