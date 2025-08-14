package server

import "upload-file/internal/model/health"

func (s *Server) migrate() {
    health.AutoMigrate(s.DB.GetDB())
}
