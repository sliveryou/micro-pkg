package xdb

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var errUnsupportedType = errors.New("unsupported database type")

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

// GetCheckSQL 获取检查数据库状态 SQL 语句
func (t Type) GetCheckSQL() (checkSQL string) {
	return "SELECT 1"
}

// GetVersionSQL 获取查询数据库版本 SQL 语句
func (t Type) GetVersionSQL() (versionSQL string) {
	switch t {
	case MySQL, PostgreSQL:
		versionSQL = "SELECT VERSION()"
	case SQLite:
		versionSQL = "SELECT SQLITE_VERSION()"
	case SQLServer:
		versionSQL = "SELECT @@VERSION"
	}

	return
}

// GetCreateSQL 获取创建数据库 SQL 语句
func (t Type) GetCreateSQL(database string) (createSQL string) {
	switch t {
	case MySQL:
		createSQL = "CREATE DATABASE IF NOT EXISTS " + database + " CHARACTER SET utf8mb4"
	case PostgreSQL:
		createSQL = "CREATE DATABASE " + database + " WITH ENCODING='UTF8'"
	case SQLServer:
		createSQL = "CREATE DATABASE " + database + " COLLATE Latin1_General_100_CI_AS_SC_UTF8"
	}

	return
}

// CheckDB 检查数据库状态
func (t Type) CheckDB(db *gorm.DB) error {
	checkSQL := t.GetCheckSQL()
	if checkSQL == "" {
		return errUnsupportedType
	}

	var ok bool
	err := db.Raw(checkSQL).Scan(&ok).Error

	return err
}

// VersionDB 查询数据库版本
func (t Type) VersionDB(db *gorm.DB) (string, error) {
	versionSQL := t.GetVersionSQL()
	if versionSQL == "" {
		return "", errUnsupportedType
	}

	var version string
	err := db.Raw(versionSQL).Scan(&version).Error

	return version, err
}

// CreateDB 创建数据库
func (t Type) CreateDB(db *gorm.DB, database string) error {
	createSQL := t.GetCreateSQL(database)
	if createSQL == "" {
		return errUnsupportedType
	}

	err := db.Exec(createSQL).Error

	return err
}

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
