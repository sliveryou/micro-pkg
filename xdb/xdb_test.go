package xdb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/conf"
)

func TestLoadConfig(t *testing.T) {
	configYaml1 := `
Type: "mysql"
Host: "localhost"
Port: 3306
User: "root"
Password: "password"
Database: "database"
MaxIdleConns: 10
MaxOpenConns: 50
LogLevel: "info"
`
	c1 := Config{}
	err := conf.LoadFromYamlBytes([]byte(configYaml1), &c1)
	require.NoError(t, err)
	assert.Equal(t, time.Hour, c1.ConnMaxLifeTime)
	assert.Equal(t, time.Hour, c1.ConnMaxIdleTime)
	assert.Equal(t, 200*time.Millisecond, c1.SlowThreshold)

	configYaml2 := `
Type: "unknown"
Host: "localhost"
Port: 3306
User: "root"
Password: "password"
Database: "database"
`
	c2 := Config{}
	err = conf.LoadFromYamlBytes([]byte(configYaml2), &c2)
	require.ErrorContains(t, err, `value "unknown" is not defined in options "[mysql postgres sqlite]"`)

	c3 := Config{}
	err = c3.fillDefault()
	require.NoError(t, err)
	assert.Equal(t, time.Hour, c3.ConnMaxLifeTime)
	assert.Equal(t, time.Hour, c3.ConnMaxIdleTime)
	assert.Equal(t, 200*time.Millisecond, c3.SlowThreshold)
}

func TestSQLite(t *testing.T) {
	c := Config{
		Type:     SQLite,
		Database: "../testdata/test.db",
		LogLevel: Info,
	}

	db, err := NewDB(c)
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = sqlDB.Ping()
	require.NoError(t, err)

	var version string
	err = db.Raw("SELECT SQLITE_VERSION()").Scan(&version).Error
	require.NoError(t, err)
	t.Log(version)
}
