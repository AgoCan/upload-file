package server

import (
	"github.com/gin-gonic/gin"

	"upload-file/internal/handler/health"
	"upload-file/internal/handler/upload"
)

// SetupRouter 初始化gin入口，路由信息
func (s *Server) SetupRouter() {
	// 客户端通过变量进行传递
	v1Router := s.Gin.Group("/api/v1")
	
	// 注册处理器
	uploadHandler := upload.NewHandler(s.Config, s.DB)
	v1Router.Use(func(c *gin.Context) {
		c.Set("handler", uploadHandler)
		c.Next()
	})
	
	healthRouter(v1Router)
	uploadRouter(v1Router)
}

func healthRouter(group *gin.RouterGroup) {
	group.GET("/health", health.HealthHandler())
}

func uploadRouter(group *gin.RouterGroup) {
	group.POST("/upload/init", upload.InitUploadHandler())
	group.POST("/upload/chunk", upload.UploadChunkHandler())
	group.POST("/upload/complete", upload.CompleteUploadHandler())
	group.GET("/upload/status/:id", upload.GetUploadStatusHandler())
	group.GET("/upload/file/:id", upload.DownloadFileHandler())
	group.DELETE("/upload/file/:id", upload.DeleteFileHandler())
	group.GET("/upload/files", upload.ListFilesHandler())
}
