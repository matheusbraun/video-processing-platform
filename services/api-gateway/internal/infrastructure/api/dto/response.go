package dto

import "time"

type UploadResponse struct {
	VideoID  string `json:"video_id"`
	Filename string `json:"filename"`
	Status   string `json:"status"`
}

type VideoInfo struct {
	ID          string     `json:"id"`
	Filename    string     `json:"filename"`
	Status      string     `json:"status"`
	FrameCount  *int       `json:"frame_count"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type ListResponse struct {
	Videos  []VideoInfo `json:"videos"`
	Total   int64       `json:"total"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
	HasMore bool        `json:"has_more"`
}

type StatusResponse struct {
	VideoID      string     `json:"video_id"`
	Filename     string     `json:"filename"`
	Status       string     `json:"status"`
	FrameCount   *int       `json:"frame_count"`
	ErrorMessage *string    `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
}

type DownloadResponse struct {
	DownloadURL string `json:"download_url"`
	Filename    string `json:"filename"`
	ExpiresIn   int64  `json:"expires_in"`
}
