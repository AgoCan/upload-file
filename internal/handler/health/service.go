package health

import (
	"upload-file/internal/config"
	"upload-file/internal/pkg/database"
	"upload-file/internal/pkg/response"
	healthModel "upload-file/internal/model/health"
)

type Health struct{
	Config *config.Config
	DB     database.DB
	HealthModelClient *healthModel.Client
}

func (h *Health) Status() response.Response {
	return response.Success("health")
}
