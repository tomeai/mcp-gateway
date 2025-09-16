package model

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type McpClient struct {
	gorm.Model

	Name        string `json:"name" gorm:"uniqueIndex;not null"`
	Description string `json:"description"`

	AccessToken string `json:"access_token" gorm:"unique; not null"`

	// AllowList contains a list of MCP Server names that this client is allowed to view and call
	// storing the list of server names as a JSON array is a convenient way for now.
	// In the future, this will be removed in favor of a separate table for ACLs.
	AllowList datatypes.JSON `json:"allow_list" gorm:"type:jsonb; not null"`
}
