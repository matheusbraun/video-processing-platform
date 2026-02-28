package controller

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/video-platform/services/api-gateway/internal/controller"
	"github.com/video-platform/services/api-gateway/internal/presenter"
	"github.com/video-platform/services/api-gateway/internal/usecase/commands"
	"github.com/video-platform/shared/pkg/auth/jwt"
	"github.com/video-platform/shared/pkg/rest"
)

type VideoHTTPController struct {
	controller controller.VideoController
	presenter  presenter.VideoPresenter
}

func NewVideoHTTPController(
	controller controller.VideoController,
	presenter presenter.VideoPresenter,
) *VideoHTTPController {
	return &VideoHTTPController{
		controller: controller,
		presenter:  presenter,
	}
}

func (h *VideoHTTPController) RegisterRoutes(r chi.Router, jwtManager jwt.JWTManager) {
	r.Post("/videos/upload", jwt.Middleware(jwtManager)(http.HandlerFunc(h.Upload)).ServeHTTP)
	r.Get("/videos", jwt.Middleware(jwtManager)(http.HandlerFunc(h.List)).ServeHTTP)
	r.Get("/videos/{id}/status", jwt.Middleware(jwtManager)(http.HandlerFunc(h.Status)).ServeHTTP)
	r.Get("/videos/{id}/download", jwt.Middleware(jwtManager)(http.HandlerFunc(h.Download)).ServeHTTP)
}

// Upload godoc
// @Summary Upload a video for processing
// @Description Uploads a video file for asynchronous frame extraction processing (1 FPS)
// @Tags videos
// @Accept multipart/form-data
// @Produce json
// @Param video formData file true "Video file to upload (max 500MB)"
// @Success 201 {object} dto.UploadResponse "Video uploaded successfully"
// @Failure 400 {object} rest.ErrorResponse "Invalid request or file validation failed"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized - missing or invalid JWT token"
// @Security BearerAuth
// @Router /videos/upload [post]
func (h *VideoHTTPController) Upload(w http.ResponseWriter, r *http.Request) {
	claims, ok := jwt.GetClaimsFromContext(r.Context())
	if !ok {
		rest.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	if err := r.ParseMultipartForm(500 << 20); err != nil {
		rest.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "failed to parse form")
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		rest.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "missing video file")
		return
	}
	defer file.Close()

	cmd := commands.UploadCommand{
		UserID:      claims.UserID,
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		FileSize:    header.Size,
		FileReader:  file,
	}

	output, err := h.controller.Upload(r.Context(), cmd)
	if err != nil {
		rest.RespondError(w, http.StatusBadRequest, "UPLOAD_FAILED", err.Error())
		return
	}

	response := h.presenter.PresentUpload(output)
	rest.RespondCreated(w, response)
}

// List godoc
// @Summary List all videos for authenticated user
// @Description Get a paginated list of all videos uploaded by the authenticated user
// @Tags videos
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of videos to return (1-100, default: 20)"
// @Param offset query int false "Number of videos to skip for pagination (default: 0)"
// @Success 200 {object} dto.ListResponse "List of videos retrieved successfully"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized - missing or invalid JWT token"
// @Failure 500 {object} rest.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /videos [get]
func (h *VideoHTTPController) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := jwt.GetClaimsFromContext(r.Context())
	if !ok {
		rest.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	cmd := commands.ListCommand{
		UserID: claims.UserID,
		Limit:  limit,
		Offset: offset,
	}

	output, err := h.controller.List(r.Context(), cmd)
	if err != nil {
		rest.RespondError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	response := h.presenter.PresentList(output)
	rest.RespondSuccess(w, response)
}

// Status godoc
// @Summary Get video processing status
// @Description Retrieve detailed status information for a specific video including processing state, frame count, and timestamps
// @Tags videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID (UUID format)"
// @Success 200 {object} dto.StatusResponse "Video status retrieved successfully"
// @Failure 400 {object} rest.ErrorResponse "Invalid video ID format"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized - missing or invalid JWT token"
// @Failure 404 {object} rest.ErrorResponse "Video not found or does not belong to user"
// @Security BearerAuth
// @Router /videos/{id}/status [get]
func (h *VideoHTTPController) Status(w http.ResponseWriter, r *http.Request) {
	claims, ok := jwt.GetClaimsFromContext(r.Context())
	if !ok {
		rest.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	videoIDStr := chi.URLParam(r, "id")
	videoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		rest.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid video ID")
		return
	}

	cmd := commands.StatusCommand{
		VideoID: videoID,
		UserID:  claims.UserID,
	}

	output, err := h.controller.Status(r.Context(), cmd)
	if err != nil {
		rest.RespondError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	response := h.presenter.PresentStatus(output)
	rest.RespondSuccess(w, response)
}

// Download godoc
// @Summary Get download URL for processed frames
// @Description Generate a presigned S3 URL to download the ZIP file containing extracted video frames
// @Tags videos
// @Accept json
// @Produce json
// @Param id path string true "Video ID (UUID format)"
// @Success 200 {object} dto.DownloadResponse "Download URL generated successfully"
// @Failure 400 {object} rest.ErrorResponse "Invalid video ID or video processing not completed"
// @Failure 401 {object} rest.ErrorResponse "Unauthorized - missing or invalid JWT token"
// @Failure 404 {object} rest.ErrorResponse "Video not found or does not belong to user"
// @Security BearerAuth
// @Router /videos/{id}/download [get]
func (h *VideoHTTPController) Download(w http.ResponseWriter, r *http.Request) {
	claims, ok := jwt.GetClaimsFromContext(r.Context())
	if !ok {
		rest.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authentication")
		return
	}

	videoIDStr := chi.URLParam(r, "id")
	videoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		rest.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid video ID")
		return
	}

	cmd := commands.DownloadCommand{
		VideoID: videoID,
		UserID:  claims.UserID,
	}

	output, err := h.controller.Download(r.Context(), cmd)
	if err != nil {
		rest.RespondError(w, http.StatusBadRequest, "DOWNLOAD_FAILED", err.Error())
		return
	}

	response := h.presenter.PresentDownload(output)
	rest.RespondSuccess(w, response)
}
