package presenter

import (
	"github.com/video-platform/services/api-gateway/internal/infrastructure/api/dto"
	"github.com/video-platform/services/api-gateway/internal/usecase/download"
	"github.com/video-platform/services/api-gateway/internal/usecase/list"
	"github.com/video-platform/services/api-gateway/internal/usecase/status"
	"github.com/video-platform/services/api-gateway/internal/usecase/upload"
)

type videoPresenterImpl struct{}

func NewVideoPresenter() VideoPresenter {
	return &videoPresenterImpl{}
}

func (p *videoPresenterImpl) PresentUpload(output *upload.UploadOutput) *dto.UploadResponse {
	return &dto.UploadResponse{
		VideoID:  output.VideoID.String(),
		Filename: output.Filename,
		Status:   output.Status,
	}
}

func (p *videoPresenterImpl) PresentList(output *list.ListOutput) *dto.ListResponse {
	videos := make([]dto.VideoInfo, len(output.Videos))
	for i, v := range output.Videos {
		videos[i] = dto.VideoInfo{
			ID:          v.ID.String(),
			Filename:    v.Filename,
			Status:      v.Status,
			FrameCount:  v.FrameCount,
			CreatedAt:   v.CreatedAt,
			CompletedAt: v.CompletedAt,
		}
	}

	return &dto.ListResponse{
		Videos:  videos,
		Total:   output.Total,
		Limit:   output.Limit,
		Offset:  output.Offset,
		HasMore: output.HasMore,
	}
}

func (p *videoPresenterImpl) PresentStatus(output *status.StatusOutput) *dto.StatusResponse {
	return &dto.StatusResponse{
		VideoID:      output.VideoID.String(),
		Filename:     output.Filename,
		Status:       output.Status,
		FrameCount:   output.FrameCount,
		ErrorMessage: output.ErrorMessage,
		CreatedAt:    output.CreatedAt,
		StartedAt:    output.StartedAt,
		CompletedAt:  output.CompletedAt,
	}
}

func (p *videoPresenterImpl) PresentDownload(output *download.DownloadOutput) *dto.DownloadResponse {
	return &dto.DownloadResponse{
		DownloadURL: output.DownloadURL,
		Filename:    output.Filename,
		ExpiresIn:   output.ExpiresIn,
	}
}
