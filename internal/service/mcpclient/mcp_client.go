package mcpclient

import (
	"errors"
	"github.com/tomeai/mcp-gateway/internal/model"
	"gorm.io/gorm"
)

type McpClientService struct {
	db *gorm.DB
}

func NewMCPClientService(db *gorm.DB) *McpClientService {
	return &McpClientService{db: db}
}

// GetClientByToken retrieves an MCP client by its access token from the database.
// It returns an error if no such client is found.
func (m *McpClientService) GetClientByToken(token string) (*model.McpClient, error) {
	var client model.McpClient
	if err := m.db.Where("access_token = ?", token).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("client not found")
		}
		return nil, err
	}
	return &client, nil
}
