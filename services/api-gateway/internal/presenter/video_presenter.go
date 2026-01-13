package presenter

import (
	"github.com/video-platform/services/api-gateway/internal/infrastructure/api/dto"
	"github.com/video-platform/services/api-gateway/internal/usecase/download"
	"github.com/video-platform/services/api-gateway/internal/usecase/list"
	"github.com/video-platform/services/api-gateway/internal/usecase/status"
	"github.com/video-platform/services/api-gateway/internal/usecase/upload"
)

type VideoPresenter interface {
	PresentUpload(output *upload.UploadOutput) *dto.UploadResponse
	PresentList(output *list.ListOutput) *dto.ListResponse
	PresentStatus(output *status.StatusOutput) *dto.StatusResponse
	PresentDownload(output *download.DownloadOutput) *dto.DownloadResponse
}
