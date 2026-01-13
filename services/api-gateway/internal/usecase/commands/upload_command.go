package commands

import "io"

type UploadCommand struct {
	UserID      int64
	Filename    string
	ContentType string
	FileSize    int64
	FileReader  io.Reader
}
