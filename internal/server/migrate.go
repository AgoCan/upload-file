package server

import (
	"upload-file/internal/model/health"
	"upload-file/internal/model/upload"
)

func (s *Server) migrate() {
    health.AutoMigrate(s.DB.GetDB())
    upload.AutoMigrate(s.DB.GetDB())
}
