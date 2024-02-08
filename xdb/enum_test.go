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
