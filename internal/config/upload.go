package config

type Upload struct {
	UploadDir     string `mapstructure:"upload_dir"`     // 上传文件存储目录
	TempDir       string `mapstructure:"temp_dir"`       // 临时文件存储目录
	MaxFileSize   int64  `mapstructure:"max_file_size"`  // 最大文件大小（字节）
	ChunkSize     int64  `mapstructure:"chunk_size"`     // 默认分片大小（字节）
	AllowedTypes  string `mapstructure:"allowed_types"`  // 允许的文件类型，逗号分隔
	CleanupExpiry int    `mapstructure:"cleanup_expiry"` // 未完成上传的文件清理时间（小时）
}