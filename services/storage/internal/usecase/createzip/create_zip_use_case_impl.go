package createzip

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/video-platform/services/storage/internal/usecase/commands"
	"github.com/video-platform/shared/pkg/logging"
	"github.com/video-platform/shared/pkg/storage/s3"
)

type createZipUseCaseImpl struct {
	s3Client s3.S3Client
}

func NewCreateZipUseCase(s3Client s3.S3Client) CreateZipUseCase {
	return &createZipUseCaseImpl{
		s3Client: s3Client,
	}
}

func (uc *createZipUseCaseImpl) Execute(ctx context.Context, cmd commands.CreateZipCommand) (*CreateZipOutput, error) {
	logging.Info("Creating ZIP file", "video_id", cmd.VideoID, "prefix", cmd.S3Prefix)

	files, err := uc.s3Client.ListObjects(ctx, "", cmd.S3Prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list frames: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no frames found for prefix: %s", cmd.S3Prefix)
	}

	logging.Info("Found frames", "count", len(files))

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	for _, fileKey := range files {
		reader, err := uc.s3Client.GetObject(ctx, "", fileKey)
		if err != nil {
			zipWriter.Close()
			return nil, fmt.Errorf("failed to get frame %s: %w", fileKey, err)
		}

		fileName := filepath.Base(fileKey)
		writer, err := zipWriter.Create(fileName)
		if err != nil {
			reader.Close()
			zipWriter.Close()
			return nil, fmt.Errorf("failed to create zip entry %s: %w", fileName, err)
		}

		if _, err := io.Copy(writer, reader); err != nil {
			reader.Close()
			zipWriter.Close()
			return nil, fmt.Errorf("failed to write frame to zip %s: %w", fileName, err)
		}
		reader.Close()
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	zipBytes := buf.Bytes()
	zipSize := int64(len(zipBytes))

	logging.Info("Uploading ZIP to S3", "size_bytes", zipSize)

	if err := uc.s3Client.Upload(ctx, "", cmd.OutputKey, bytes.NewReader(zipBytes)); err != nil {
		return nil, fmt.Errorf("failed to upload zip: %w", err)
	}

	logging.Info("ZIP created successfully", "output_key", cmd.OutputKey, "file_count", len(files))

	return &CreateZipOutput{
		ZipPath:      cmd.OutputKey,
		FileCount:    len(files),
		ZipSizeBytes: zipSize,
	}, nil
}
