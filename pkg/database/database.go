package database

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config 数据库配置
type Config struct {
	Type     string // mysql, postgres, sqlite
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Options  map[string]string
}

// Database 数据库接口
type Database interface {
	Connect() error
	Close() error
	DB() *gorm.DB
}

// MySQL MySQL 数据库
type MySQL struct {
	config *Config
	db     *gorm.DB
}

// NewMySQL 创建 MySQL 数据库实例
func NewMySQL(config *Config) *MySQL {
	return &MySQL{config: config}
}

// Connect 连接数据库
func (m *MySQL) Connect() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		m.config.User,
		m.config.Password,
		m.config.Host,
		m.config.Port,
		m.config.Database,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	m.db = db
	return nil
}

// Close 关闭数据库连接
func (m *MySQL) Close() error {
	if m.db != nil {
		sqlDB, err := m.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// DB 获取数据库实例
func (m *MySQL) DB() *gorm.DB {
	return m.db
}

// PostgreSQL PostgreSQL 数据库
type PostgreSQL struct {
	config *Config
	db     *gorm.DB
}

// NewPostgreSQL 创建 PostgreSQL 数据库实例
func NewPostgreSQL(config *Config) *PostgreSQL {
	return &PostgreSQL{config: config}
}

// Connect 连接数据库
func (p *PostgreSQL) Connect() error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.config.Host,
		p.config.Port,
		p.config.User,
		p.config.Password,
		p.config.Database,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	p.db = db
	return nil
}

// Close 关闭数据库连接
func (p *PostgreSQL) Close() error {
	if p.db != nil {
		sqlDB, err := p.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// DB 获取数据库实例
func (p *PostgreSQL) DB() *gorm.DB {
	return p.db
}

// SQLite SQLite 数据库
type SQLite struct {
	config *Config
	db     *gorm.DB
}

// NewSQLite 创建 SQLite 数据库实例
func NewSQLite(config *Config) *SQLite {
	return &SQLite{config: config}
}

// Connect 连接数据库
func (s *SQLite) Connect() error {
	db, err := gorm.Open(sqlite.Open(s.config.Database), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	s.db = db
	return nil
}

// Close 关闭数据库连接
func (s *SQLite) Close() error {
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// DB 获取数据库实例
func (s *SQLite) DB() *gorm.DB {
	return s.db
}

// NewDatabase 创建数据库实例
func NewDatabase(config *Config) (Database, error) {
	switch config.Type {
	case "mysql":
		return NewMySQL(config), nil
	case "postgres":
		return NewPostgreSQL(config), nil
	case "sqlite":
		return NewSQLite(config), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}
