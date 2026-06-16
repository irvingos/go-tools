package selectdb

import (
	"fmt"
	"time"

	"github.com/irvingos/go-tools/logx"
	"gorm.io/driver/mysql"
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
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Asia%%2FShanghai&timeout=5s&readTimeout=120s&writeTimeout=120s",
		c.User, c.Password, c.Host, c.Port, c.DBName,
	)
}

func New(cfg Config) (db *gorm.DB, err error) {
	db, err = gorm.Open(
		mysql.Open(cfg.buildDSN()),
		&gorm.Config{
			Logger: logx.NewDBLogger(&logx.DBLoggerOptions{
				SlowSQLThreshold: cfg.SlowSQLThreshold,
			}).LogMode(cfg.LogLevel),
			PrepareStmt:            false,
			SkipDefaultTransaction: true,
		})
	if err != nil {
		return db, fmt.Errorf("error when init selectdb, err: %v", err)
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
