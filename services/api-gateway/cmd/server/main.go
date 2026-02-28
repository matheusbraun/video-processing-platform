package main

import (
	"github.com/video-platform/services/api-gateway/internal/app"

	_ "github.com/video-platform/services/api-gateway/docs"
)

// @title Video Processing Platform API
// @version 1.0
// @description Microservices platform for video frame extraction at 1 FPS with asynchronous processing
// @description This API allows users to upload videos, track processing status, and download extracted frames as ZIP files.

// @contact.name API Support
// @contact.email support@video-platform.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.

func main() {
	app.InitializeApp().Run()
}
