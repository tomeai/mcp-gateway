package db

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"github.com/urfave/cli/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

func NewDBConnection(ctx *cli.Context) (*gorm.DB, error) {
	var dialector gorm.Dialector
	dsn := ctx.String("dsn")
	if dsn == "" {
		log.Println("[db] DATABASE_URL not set – falling back to embedded SQLite ./mcp.db")
		dialector = sqlite.Open("mcp.db?_busy_timeout=5000&_journal_mode=WAL")
	} else {
		dialector = postgres.Open(dsn)
	}

	c := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}
	db, err := gorm.Open(dialector, c)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}
