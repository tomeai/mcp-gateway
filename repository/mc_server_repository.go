package repository

import (
	"github.com/tomeai/mcp-gateway/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type McpServerService struct {
	db *gorm.DB
}

func NewMcpServerService(db *gorm.DB) *McpServerService {
	return &McpServerService{db: db}
}

func (ms *McpServerService) UpsertMcpServer(server *model.McpServer) error {
	return ms.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "server_name"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"server_config": server.ServerConfig,
			"updated_at":    gorm.Expr("NOW()"),
		}),
	}).Create(server).Error
}

func (ms *McpServerService) GetMcpServer(userId, serverName string) (*model.McpServer, error) {
	var server model.McpServer
	err := ms.db.Where("user_id = ? AND server_name = ?", userId, serverName).First(&server).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}
