package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/video-platform/services/storage/internal/controller"
	"github.com/video-platform/services/storage/internal/usecase/commands"
	"github.com/video-platform/shared/pkg/rest"
)

type CreateZipRequest struct {
	VideoID   string `json:"video_id"`
	S3Prefix  string `json:"s3_prefix"`
	OutputKey string `json:"output_key"`
}

type StorageHTTPController struct {
	controller controller.StorageController
}

func NewStorageHTTPController(controller controller.StorageController) *StorageHTTPController {
	return &StorageHTTPController{
		controller: controller,
	}
}

func (h *StorageHTTPController) RegisterRoutes(r chi.Router) {
	r.Post("/internal/zip/create", h.CreateZip)
}

func (h *StorageHTTPController) CreateZip(w http.ResponseWriter, r *http.Request) {
	var req CreateZipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	cmd := commands.CreateZipCommand{
		VideoID:   req.VideoID,
		S3Prefix:  req.S3Prefix,
		OutputKey: req.OutputKey,
	}

	output, err := h.controller.CreateZip(r.Context(), cmd)
	if err != nil {
		rest.RespondError(w, http.StatusInternalServerError, "ZIP_CREATION_FAILED", err.Error())
		return
	}

	rest.RespondSuccess(w, output)
}
