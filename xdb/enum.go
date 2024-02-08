package xdb

import (
	"gorm.io/gorm/logger"
)

// Type 数据库类型
type Type string

const (
	// MySQL 数据库类型：mysql
	MySQL Type = "mysql"
	// PostgreSQL 数据库类型：postgres
	PostgreSQL Type = "postgres"
	// SQLite 数据库类型：sqlite
	SQLite Type = "sqlite"
	// SQLServer 数据库类型：sqlserver
	SQLServer Type = "sqlserver"
)

// LogLevel 日志级别
type LogLevel string

const (
	// Info 日志级别：info
	Info LogLevel = "info"
	// Warn 日志级别：warn
	Warn LogLevel = "warn"
	// Error 日志级别：error
	Error LogLevel = "error"
	// Silent 日志级别：silent
	Silent LogLevel = "silent"
)

// ToGROMLogLevel 转化日志级别
func (l LogLevel) ToGROMLogLevel() (logLevel logger.LogLevel) {
	switch l {
	case Info:
		logLevel = logger.Info
	case Warn:
		logLevel = logger.Warn
	case Error:
		logLevel = logger.Error
	case Silent:
		logLevel = logger.Silent
	default:
		logLevel = logger.Info
	}

	return
}
