package upload

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"upload-file/internal/config"
	"upload-file/internal/pkg/database"
	"upload-file/internal/pkg/response"
)

// Handler 文件上传处理器
type Handler struct {
	Config       *config.Config
	DB           database.DB
	UploadClient *Service
}

// NewHandler 创建文件上传处理器
func NewHandler(config *config.Config, db database.DB) *Handler {
	// 确保上传目录存在
	os.MkdirAll(config.Upload.UploadDir, 0755)
	os.MkdirAll(config.Upload.TempDir, 0755)

	return &Handler{
		Config:       config,
		DB:           db,
		UploadClient: NewService(config, db),
	}
}

// InitUploadHandler 初始化上传，获取上传ID和分片信息
func InitUploadHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := c.MustGet("handler").(*Handler)

		var req InitUploadRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.Error(response.ErrCodeParameter))
			return
		}

		// 验证文件大小
		if req.FileSize > handler.Config.Upload.MaxFileSize {
			c.JSON(http.StatusBadRequest, response.ErrorUnknown(response.ErrCodeParameter, "文件大小超过限制"))
			return
		}

		// 初始化上传
		result, err := handler.UploadClient.InitUpload(&req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorUnknown(response.ErrSQL, err.Error()))
			return
		}

		c.JSON(http.StatusOK, response.Success(result))
	}
}

// UploadChunkHandler 上传文件分片
func UploadChunkHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := c.MustGet("handler").(*Handler)

		// 获取参数
		fileID, err := strconv.ParseUint(c.PostForm("file_id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.Error(response.ErrCodeParameter))
			return
		}

		chunkNum, err := strconv.Atoi(c.PostForm("chunk_num"))
		if err != nil {
			c.JSON(http.StatusBadRequest, response.Error(response.ErrCodeParameter))
			return
		}

		// 获取上传的文件
		file, err := c.FormFile("chunk")
		if err != nil {
			c.JSON(http.StatusBadRequest, response.ErrorUnknown(response.ErrCodeParameter, "无法获取上传文件"))
			return
		}

		// 保存分片
		err = handler.UploadClient.SaveChunk(uint(fileID), chunkNum, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorUnknown(response.ErrSQL, err.Error()))
			return
		}

		c.JSON(http.StatusOK, response.Success(gin.H{
			"file_id":   fileID,
			"chunk_num": chunkNum,
			"status":    "success",
		}))
	}
}

// CompleteUploadHandler 完成上传，合并分片
func CompleteUploadHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := c.MustGet("handler").(*Handler)

		var req CompleteUploadRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.Error(response.ErrCodeParameter))
			return
		}

		// 合并分片
		fileInfo, err := handler.UploadClient.CompleteUpload(req.FileID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorUnknown(response.ErrSQL, err.Error()))
			return
		}

		c.JSON(http.StatusOK, response.Success(gin.H{
			"file_id":   fileInfo.ID,
			"file_name": fileInfo.FileName,
			"file_path": fileInfo.FilePath,
			"file_size": fileInfo.FileSize,
			"status":    fileInfo.Status,
		}))
	}
}

// GetUploadStatusHandler 获取上传状态
func GetUploadStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := c.MustGet("handler").(*Handler)

		fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.Error(response.ErrCodeParameter))
			return
		}

		// 获取上传状态
		status, err := handler.UploadClient.GetUploadStatus(uint(fileID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorUnknown(response.ErrSQL, err.Error()))
			return
		}

		c.JSON(http.StatusOK, response.Success(status))
	}
}

// DownloadFileHandler 下载文件
func DownloadFileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := c.MustGet("handler").(*Handler)

		fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.Error(response.ErrCodeParameter))
			return
		}

		// 获取文件信息
		fileInfo, err := handler.UploadClient.GetFileInfo(uint(fileID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorUnknown(response.ErrSQL, err.Error()))
			return
		}

		// 检查文件是否存在
		if _, err := os.Stat(fileInfo.FilePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, response.ErrorUnknown(response.ErrCodeParameter, "文件不存在"))
			return
		}

		// 设置文件名
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.FileName))
		c.Header("Content-Type", fileInfo.ContentType)
		c.File(fileInfo.FilePath)
	}
}

// DeleteFileHandler 删除文件
func DeleteFileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := c.MustGet("handler").(*Handler)

		fileID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.Error(response.ErrCodeParameter))
			return
		}

		// 删除文件
		err = handler.UploadClient.DeleteFile(uint(fileID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorUnknown(response.ErrSQL, err.Error()))
			return
		}

		c.JSON(http.StatusOK, response.Success(gin.H{
			"file_id": fileID,
			"status":  "deleted",
		}))
	}
}

// ListFilesHandler 列出所有文件
func ListFilesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := c.MustGet("handler").(*Handler)

		// 获取文件列表
		files, err := handler.UploadClient.ListFiles()
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorUnknown(response.ErrSQL, err.Error()))
			return
		}

		c.JSON(http.StatusOK, response.Success(files))
	}
}
