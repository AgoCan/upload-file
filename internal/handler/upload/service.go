package upload

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"upload-file/internal/config"
	"upload-file/internal/model/upload"
	"upload-file/internal/pkg/database"
)

// Service 文件上传服务
type Service struct {
	Config      *config.Config
	DB          database.DB
	UploadModel *upload.Client
}

// NewService 创建文件上传服务
func NewService(config *config.Config, db database.DB) *Service {
	return &Service{
		Config:      config,
		DB:          db,
		UploadModel: upload.NewClient(db.GetDB()),
	}
}

// InitUpload 初始化上传
func (s *Service) InitUpload(req *InitUploadRequest) (*InitUploadResponse, error) {
	// 检查文件是否已存在（通过哈希值）
	existingFile, err := s.UploadModel.GetFileInfoByHash(req.FileHash)
	if err == nil && existingFile.ID > 0 && existingFile.Status == "completed" {
		// 文件已存在，直接返回
		chunks, _ := s.UploadModel.GetChunksByFileID(existingFile.ID)
		uploadedChunks := make([]int, 0)
		for _, chunk := range chunks {
			uploadedChunks = append(uploadedChunks, chunk.ChunkNum)
		}

		totalChunks := int((req.FileSize + s.Config.Upload.ChunkSize - 1) / s.Config.Upload.ChunkSize)
		return &InitUploadResponse{
			FileID:      existingFile.ID,
			UploadID:    uuid.New().String(),
			ChunkSize:   s.Config.Upload.ChunkSize,
			TotalChunks: totalChunks,
			Status:      existingFile.Status,
			Uploaded:    uploadedChunks,
		}, nil
	}

	// 创建新的文件记录
	fileInfo := &upload.FileInfo{
		FileName:    req.FileName,
		FilePath:    filepath.Join(s.Config.Upload.UploadDir, req.FileHash),
		FileSize:    req.FileSize,
		FileHash:    req.FileHash,
		ContentType: req.ContentType,
		Status:      "uploading",
	}

	err = s.UploadModel.CreateFileInfo(fileInfo)
	if err != nil {
		return nil, err
	}

	// 计算分片数量
	totalChunks := int((req.FileSize + s.Config.Upload.ChunkSize - 1) / s.Config.Upload.ChunkSize)

	return &InitUploadResponse{
		FileID:      fileInfo.ID,
		UploadID:    uuid.New().String(),
		ChunkSize:   s.Config.Upload.ChunkSize,
		TotalChunks: totalChunks,
		Status:      "initialized",
		Uploaded:    []int{},
	}, nil
}

// SaveChunk 保存分片
func (s *Service) SaveChunk(fileID uint, chunkNum int, file *multipart.FileHeader) error {
	// 获取文件信息
	var fileInfo upload.FileInfo
	if err := s.DB.GetDB().First(&fileInfo, fileID).Error; err != nil {
		return err
	}

	// 创建临时目录
	chunkDir := filepath.Join(s.Config.Upload.TempDir, strconv.FormatUint(uint64(fileID), 10))
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		return err
	}

	// 保存分片文件
	chunkPath := filepath.Join(chunkDir, fmt.Sprintf("%d", chunkNum))
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(chunkPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	// 保存分片信息
	chunkInfo := &upload.ChunkInfo{
		FileID:    fileID,
		ChunkNum:  chunkNum,
		ChunkSize: file.Size,
		ChunkPath: chunkPath,
	}

	// 检查分片是否已存在
	var existingChunk upload.ChunkInfo
	err = s.DB.GetDB().Where("file_id = ? AND chunk_num = ?", fileID, chunkNum).First(&existingChunk).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 分片不存在，创建新记录
			return s.UploadModel.CreateChunkInfo(chunkInfo)
		}
		return err
	}

	// 分片已存在，更新记录
	existingChunk.ChunkSize = file.Size
	existingChunk.ChunkPath = chunkPath
	return s.DB.GetDB().Save(&existingChunk).Error
}

// CompleteUpload 完成上传，合并分片
func (s *Service) CompleteUpload(fileID uint) (*upload.FileInfo, error) {
	// 获取文件信息
	var fileInfo upload.FileInfo
	if err := s.DB.GetDB().First(&fileInfo, fileID).Error; err != nil {
		return nil, err
	}

	// 获取所有分片
	chunks, err := s.UploadModel.GetChunksByFileID(fileID)
	if err != nil {
		return nil, err
	}

	// 计算预期的分片数量
	expectedChunks := int((fileInfo.FileSize + s.Config.Upload.ChunkSize - 1) / s.Config.Upload.ChunkSize)
	if len(chunks) != expectedChunks {
		return nil, fmt.Errorf("分片不完整，预期 %d 个分片，实际 %d 个分片", expectedChunks, len(chunks))
	}

	// 按分片序号排序
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].ChunkNum < chunks[j].ChunkNum
	})

	// 创建最终文件
	finalPath := filepath.Join(s.Config.Upload.UploadDir, fileInfo.FileHash)
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return nil, err
	}
	defer finalFile.Close()

	// 合并分片
	hash := md5.New()
	for _, chunk := range chunks {
		chunkFile, err := os.Open(chunk.ChunkPath)
		if err != nil {
			return nil, err
		}

		if _, err = io.Copy(finalFile, chunkFile); err != nil {
			chunkFile.Close()
			return nil, err
		}

		if _, err = io.Copy(hash, chunkFile); err != nil {
			chunkFile.Close()
			return nil, err
		}

		chunkFile.Close()
	}

	// 更新文件状态
	fileInfo.Status = "completed"
	fileInfo.FilePath = finalPath
	if err := s.UploadModel.UpdateFileInfo(&fileInfo); err != nil {
		return nil, err
	}

	// 清理临时文件
	chunkDir := filepath.Join(s.Config.Upload.TempDir, strconv.FormatUint(uint64(fileID), 10))
	os.RemoveAll(chunkDir)

	return &fileInfo, nil
}

// GetUploadStatus 获取上传状态
func (s *Service) GetUploadStatus(fileID uint) (*UploadStatusResponse, error) {
	// 获取文件信息
	var fileInfo upload.FileInfo
	if err := s.DB.GetDB().First(&fileInfo, fileID).Error; err != nil {
		return nil, err
	}

	// 获取已上传的分片
	chunks, err := s.UploadModel.GetChunksByFileID(fileID)
	if err != nil {
		return nil, err
	}

	// 计算总分片数
	totalChunks := int((fileInfo.FileSize + s.Config.Upload.ChunkSize - 1) / s.Config.Upload.ChunkSize)

	// 构建已上传分片列表
	uploadedChunks := make([]int, 0)
	for _, chunk := range chunks {
		uploadedChunks = append(uploadedChunks, chunk.ChunkNum)
	}

	// 计算上传进度
	progress := 0
	if totalChunks > 0 {
		progress = len(chunks) * 100 / totalChunks
	}

	return &UploadStatusResponse{
		FileID:      fileInfo.ID,
		FileName:    fileInfo.FileName,
		FileSize:    fileInfo.FileSize,
		Status:      fileInfo.Status,
		TotalChunks: totalChunks,
		Uploaded:    uploadedChunks,
		Progress:    progress,
	}, nil
}

// GetFileInfo 获取文件信息
func (s *Service) GetFileInfo(fileID uint) (*upload.FileInfo, error) {
	var fileInfo upload.FileInfo
	if err := s.DB.GetDB().First(&fileInfo, fileID).Error; err != nil {
		return nil, err
	}
	return &fileInfo, nil
}

// DeleteFile 删除文件
func (s *Service) DeleteFile(fileID uint) error {
	// 获取文件信息
	var fileInfo upload.FileInfo
	if err := s.DB.GetDB().First(&fileInfo, fileID).Error; err != nil {
		return err
	}

	// 删除物理文件
	if fileInfo.FilePath != "" {
		os.Remove(fileInfo.FilePath)
	}

	// 删除分片
	chunkDir := filepath.Join(s.Config.Upload.TempDir, strconv.FormatUint(uint64(fileID), 10))
	os.RemoveAll(chunkDir)

	// 删除数据库记录
	if err := s.UploadModel.DeleteChunks(fileID); err != nil {
		return err
	}

	return s.DB.GetDB().Delete(&fileInfo).Error
}

// ListFiles 列出所有文件
func (s *Service) ListFiles() ([]upload.FileInfo, error) {
	var files []upload.FileInfo
	if err := s.DB.GetDB().Where("status = ?", "completed").Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

// CleanupExpiredUploads 清理过期的未完成上传
func (s *Service) CleanupExpiredUploads() error {
	expiryTime := time.Now().Add(-time.Duration(s.Config.Upload.CleanupExpiry) * time.Hour)

	var expiredFiles []upload.FileInfo
	if err := s.DB.GetDB().Where("status = ? AND updated_at < ?", "uploading", expiryTime).Find(&expiredFiles).Error; err != nil {
		return err
	}

	for _, file := range expiredFiles {
		s.DeleteFile(file.ID)
	}

	return nil
}