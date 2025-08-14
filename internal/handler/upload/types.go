package upload

import (
	"upload-file/internal/model/upload"
)

// InitUploadRequest 初始化上传请求
type InitUploadRequest struct {
	FileName    string `json:"file_name" binding:"required"`    // 文件名
	FileSize    int64  `json:"file_size" binding:"required"`    // 文件大小
	FileHash    string `json:"file_hash" binding:"required"`    // 文件哈希
	ContentType string `json:"content_type" binding:"required"` // 文件类型
}

// InitUploadResponse 初始化上传响应
type InitUploadResponse struct {
	FileID      uint   `json:"file_id"`       // 文件ID
	UploadID    string `json:"upload_id"`     // 上传ID
	ChunkSize   int64  `json:"chunk_size"`    // 分片大小
	TotalChunks int    `json:"total_chunks"`  // 总分片数
	Status      string `json:"status"`        // 状态
	Uploaded    []int  `json:"uploaded"`      // 已上传的分片
}

// CompleteUploadRequest 完成上传请求
type CompleteUploadRequest struct {
	FileID uint `json:"file_id" binding:"required"` // 文件ID
}

// UploadStatusResponse 上传状态响应
type UploadStatusResponse struct {
	FileID      uint   `json:"file_id"`       // 文件ID
	FileName    string `json:"file_name"`     // 文件名
	FileSize    int64  `json:"file_size"`     // 文件大小
	Status      string `json:"status"`        // 状态
	TotalChunks int    `json:"total_chunks"`  // 总分片数
	Uploaded    []int  `json:"uploaded"`      // 已上传的分片
	Progress    int    `json:"progress"`      // 上传进度（百分比）
}

// FileListResponse 文件列表响应
type FileListResponse struct {
	Files []upload.FileInfo `json:"files"` // 文件列表
}