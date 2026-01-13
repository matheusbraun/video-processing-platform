package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type FFmpegService interface {
	ExtractFrames(ctx context.Context, videoPath, outputDir string, fps int) (int, error)
}

type ffmpegService struct{}

func NewFFmpegService() FFmpegService {
	return &ffmpegService{}
}

func (s *ffmpegService) ExtractFrames(ctx context.Context, videoPath, outputDir string, fps int) (int, error) {
	outputPattern := filepath.Join(outputDir, "frame_%04d.jpg")

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", videoPath,
		"-vf", fmt.Sprintf("fps=%d", fps),
		"-qscale:v", "2",
		outputPattern,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffmpeg failed: %w, output: %s", err, string(output))
	}

	files, err := os.ReadDir(outputDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read output directory: %w", err)
	}

	frameCount := 0
	for _, file := range files {
		if !file.IsDir() {
			frameCount++
		}
	}

	return frameCount, nil
}

func GetVideoDuration(ctx context.Context, videoPath string) (float64, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	duration, err := strconv.ParseFloat(string(output), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}
