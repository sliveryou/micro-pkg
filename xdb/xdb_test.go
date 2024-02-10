package xdb

import (
	"os"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
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
	require.ErrorContains(t, err, `value "unknown" is not defined in options "[mysql postgres sqlite sqlserver]"`)

	c3 := Config{}
	err = c3.fillDefault()
	require.NoError(t, err)
	assert.Equal(t, time.Hour, c3.ConnMaxLifeTime)
	assert.Equal(t, time.Hour, c3.ConnMaxIdleTime)
	assert.Equal(t, 200*time.Millisecond, c3.SlowThreshold)
}

func TestMustNew(t *testing.T) {
	c := Config{
		Type:     SQLite,
		Database: "../testdata/test.db",
		LogLevel: Info,
	}
	db := MustNewDB(c)
	assert.NotNil(t, db)

	db, mock := MustNewDBMock(c)
	assert.NotNil(t, db)
	assert.NotNil(t, mock)
}

func TestSQLite(t *testing.T) {
	c := Config{
		Type:     SQLite,
		Database: "../testdata/test.db",
		LogLevel: Info,
	}

	db, err := NewDB(c)
	require.NoError(t, err)
	defer func() {
		os.Remove(c.Database + "-shm")
		os.Remove(c.Database + "-wal")
	}()

	// check db
	err = c.Type.CheckDB(db)
	require.NoError(t, err)
	// get db version
	version, err := c.Type.VersionDB(db)
	require.NoError(t, err)
	require.NotEmpty(t, version)
	t.Log(version)
}

func TestMySQL(t *testing.T) {
	c := Config{
		Type:       MySQL,
		Host:       "localhost",
		Port:       3306,
		User:       "root",
		Password:   "Abc123456",
		Database:   "my_test_db",
		NeedCreate: true,
		LogLevel:   Info,
	}
	// db, err := NewDB(c)
	db, mock, err := NewDBMock(c)
	require.NoError(t, err)

	mock.ExpectPrepare("^CREATE DATABASE IF NOT EXISTS (.+) CHARACTER SET utf8mb4").ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("^SELECT 1").ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"ok"}).AddRow("1"))
	mock.ExpectPrepare("^SELECT VERSION()").ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("5.7.31"))

	// create db
	_ = c.Type.CreateDB(db, c.Database)
	// check db
	err = c.Type.CheckDB(db)
	require.NoError(t, err)
	// get db version
	version, err := c.Type.VersionDB(db)
	require.NoError(t, err)
	require.NotEmpty(t, version)
	t.Log(version)
}

func TestPostgreSQL(t *testing.T) {
	c := Config{
		Type:       PostgreSQL,
		Host:       "localhost",
		Port:       5432,
		User:       "root",
		Password:   "Abc123456",
		Database:   "my_test_db",
		NeedCreate: true,
		LogLevel:   Info,
	}
	// db, err := NewDB(c)
	db, mock, err := NewDBMock(c)
	require.NoError(t, err)

	mock.ExpectPrepare("^CREATE DATABASE (.+) WITH ENCODING='UTF8'").ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("^SELECT 1").ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"ok"}).AddRow("1"))
	mock.ExpectPrepare("^SELECT VERSION()").ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("PostgreSQL 14.9 (Debian 14.9-1.pgdg120+1) on x86_64-pc-linux-gnu, compiled by gcc (Debian 12.2.0-14) 12.2.0, 64-bit"))

	// create db
	_ = c.Type.CreateDB(db, c.Database)
	// check db
	err = c.Type.CheckDB(db)
	require.NoError(t, err)
	// get db version
	version, err := c.Type.VersionDB(db)
	require.NoError(t, err)
	require.NotEmpty(t, version)
	t.Log(version)
}

func TestSQLServer(t *testing.T) {
	c := Config{
		Type:       SQLServer,
		Host:       "localhost",
		Port:       1433,
		User:       "sa",
		Password:   "Abc123456",
		Database:   "my_test_db",
		NeedCreate: true,
		LogLevel:   Info,
	}
	// db, err := NewDB(c)
	db, mock, err := NewDBMock(c)
	require.NoError(t, err)

	mock.ExpectPrepare("^CREATE DATABASE (.+) COLLATE Latin1_General_100_CI_AS_SC_UTF8").ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("^SELECT 1").ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"ok"}).AddRow("1"))
	mock.ExpectPrepare("^SELECT @@VERSION").ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("Microsoft SQL Server 2022 (RTM-CU11) (KB5032679) - 16.0.4105.2 (X64) \n\tNov 14 2023 18:33:19 \n\tCopyright (C) 2022 Microsoft Corporation\n\tDeveloper Edition (64-bit) on Linux (Ubuntu 22.04.3 LTS) <X64>"))

	// create db
	_ = c.Type.CreateDB(db, c.Database)
	// check db
	err = c.Type.CheckDB(db)
	require.NoError(t, err)
	// get db version
	version, err := c.Type.VersionDB(db)
	require.NoError(t, err)
	require.NotEmpty(t, version)
	t.Log(version)
}
