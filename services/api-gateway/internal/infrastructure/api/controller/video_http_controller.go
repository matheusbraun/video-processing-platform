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
