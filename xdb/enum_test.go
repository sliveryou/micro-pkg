package xdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/logger"
)

func TestLogLevel_ToGROMLogLevel(t *testing.T) {
	assert.Equal(t, logger.Info, Info.ToGROMLogLevel())
	assert.Equal(t, logger.Warn, Warn.ToGROMLogLevel())
	assert.Equal(t, logger.Error, Error.ToGROMLogLevel())
	assert.Equal(t, logger.Silent, Silent.ToGROMLogLevel())
	assert.Equal(t, logger.Info, LogLevel("unknown").ToGROMLogLevel())
}

func TestType_GetCheckSQL(t *testing.T) {
	assert.Equal(t, "SELECT 1", MySQL.GetCheckSQL())
	assert.Equal(t, "SELECT 1", PostgreSQL.GetCheckSQL())
	assert.Equal(t, "SELECT 1", SQLite.GetCheckSQL())
	assert.Equal(t, "SELECT 1", SQLServer.GetCheckSQL())
	assert.Equal(t, "SELECT 1", Type("unknown").GetCheckSQL())
}

func TestType_GetVersionSQL(t *testing.T) {
	assert.Equal(t, "SELECT VERSION()", MySQL.GetVersionSQL())
	assert.Equal(t, "SELECT VERSION()", PostgreSQL.GetVersionSQL())
	assert.Equal(t, "SELECT SQLITE_VERSION()", SQLite.GetVersionSQL())
	assert.Equal(t, "SELECT @@VERSION", SQLServer.GetVersionSQL())
	assert.Empty(t, Type("unknown").GetVersionSQL())
}

func TestType_GetCreateSQL(t *testing.T) {
	assert.Equal(t, "CREATE DATABASE IF NOT EXISTS my_test CHARACTER SET utf8mb4", MySQL.GetCreateSQL("my_test"))
	assert.Equal(t, "CREATE DATABASE my_test WITH ENCODING='UTF8'", PostgreSQL.GetCreateSQL("my_test"))
	assert.Empty(t, SQLite.GetCreateSQL("my_test"))
	assert.Equal(t, "CREATE DATABASE my_test COLLATE Latin1_General_100_CI_AS_SC_UTF8", SQLServer.GetCreateSQL("my_test"))
	assert.Empty(t, Type("unknown").GetCreateSQL("my_test"))
}
