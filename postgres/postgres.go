package postgres

import (
	"fmt"
	"time"

	"github.com/irvingos/go-tools/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host             string
	Port             int
	DBName           string
	User             string
	Password         string
	LogLevel         logger.LogLevel
	SlowSQLThreshold time.Duration
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  time.Duration
}

func (c Config) maxOpenConns() int {
	if c.MaxOpenConns <= 0 {
		return 20
	}
	return c.MaxOpenConns
}

func (c Config) maxIdleConns() int {
	if c.MaxIdleConns <= 0 {
		return 5
	}
	return c.MaxIdleConns
}

func (c Config) connMaxLifetime() time.Duration {
	if c.ConnMaxLifetime <= 0 {
		return 30 * time.Minute
	}
	return c.ConnMaxLifetime
}

func (c Config) buildDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DBName)
}

func New(cfg Config) (db *gorm.DB, err error) {
	db, err = gorm.Open(
		postgres.Open(cfg.buildDSN()),
		&gorm.Config{
			Logger: logx.NewDBLogger(&logx.DBLoggerOptions{
				SlowSQLThreshold: cfg.SlowSQLThreshold,
			}).LogMode(cfg.LogLevel),
			PrepareStmt: false,
		})
	if err != nil {
		return db, fmt.Errorf("error when init postgres, err: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.maxOpenConns())
	sqlDB.SetMaxIdleConns(cfg.maxIdleConns())
	sqlDB.SetConnMaxLifetime(cfg.connMaxLifetime())

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return
}
