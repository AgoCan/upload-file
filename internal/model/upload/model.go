package upload

import (
	"time"

	"gorm.io/gorm"
)

// FileInfo 存储文件信息的模型
type FileInfo struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	FileName    string         `gorm:"size:255;not null" json:"file_name"`      // 文件名
	FilePath    string         `gorm:"size:255;not null" json:"file_path"`      // 文件存储路径
	FileSize    int64          `gorm:"not null" json:"file_size"`               // 文件大小（字节）
	FileHash    string         `gorm:"size:64;not null;index" json:"file_hash"` // 文件哈希值，用于去重
	ContentType string         `gorm:"size:128" json:"content_type"`            // 文件MIME类型
	Status      string         `gorm:"size:20;not null" json:"status"`          // 文件状态：uploading, completed, failed
}

// ChunkInfo 存储分片信息的模型
type ChunkInfo struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	FileID    uint           `gorm:"not null;index" json:"file_id"`       // 关联的文件ID
	ChunkNum  int            `gorm:"not null" json:"chunk_num"`           // 分片序号
	ChunkSize int64          `gorm:"not null" json:"chunk_size"`          // 分片大小
	ChunkPath string         `gorm:"size:255;not null" json:"chunk_path"` // 分片存储路径
}

// Client 文件上传模型客户端
type Client struct {
	DB *gorm.DB
}

// NewClient 创建文件上传模型客户端
func NewClient(db *gorm.DB) *Client {
	return &Client{DB: db}
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&FileInfo{}, &ChunkInfo{})
}

// CreateFileInfo 创建文件信息记录
func (c *Client) CreateFileInfo(fileInfo *FileInfo) error {
	return c.DB.Create(fileInfo).Error
}

// GetFileInfoByHash 通过文件哈希获取文件信息
func (c *Client) GetFileInfoByHash(fileHash string) (*FileInfo, error) {
	var fileInfo FileInfo
	err := c.DB.Where("file_hash = ?", fileHash).First(&fileInfo).Error
	return &fileInfo, err
}

// UpdateFileInfo 更新文件信息
func (c *Client) UpdateFileInfo(fileInfo *FileInfo) error {
	return c.DB.Save(fileInfo).Error
}

// CreateChunkInfo 创建分片信息记录
func (c *Client) CreateChunkInfo(chunkInfo *ChunkInfo) error {
	return c.DB.Create(chunkInfo).Error
}

// GetChunksByFileID 获取文件的所有分片信息
func (c *Client) GetChunksByFileID(fileID uint) ([]ChunkInfo, error) {
	var chunks []ChunkInfo
	err := c.DB.Where("file_id = ?", fileID).Order("chunk_num").Find(&chunks).Error
	return chunks, err
}

// GetChunkInfo 获取特定的分片信息
func (c *Client) GetChunkInfo(fileID uint, chunkNum int) (*ChunkInfo, error) {
	var chunkInfo ChunkInfo
	err := c.DB.Where("file_id = ? AND chunk_num = ?", fileID, chunkNum).First(&chunkInfo).Error
	return &chunkInfo, err
}

// DeleteChunks 删除文件的所有分片信息
func (c *Client) DeleteChunks(fileID uint) error {
	return c.DB.Where("file_id = ?", fileID).Delete(&ChunkInfo{}).Error
}
