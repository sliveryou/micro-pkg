package dbchecker

import (
	"context"
	"encoding/json"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sliveryou/micro-pkg/xdb"
)

func TestNewChecker(t *testing.T) {
	c := xdb.Config{
		Type:     xdb.MySQL,
		Database: "my_test_db",
		LogLevel: xdb.Error,
	}
	db, _, err := xdb.NewDBMock(c)
	require.NoError(t, err)

	checker := NewChecker(c.Type, db)
	assert.NotNil(t, checker)

	assert.PanicsWithError(t, "dbchecker: nil db is invalid", func() {
		NewChecker(c.Type, nil)
	})
}

func TestChecker_Check(t *testing.T) {
	c := xdb.Config{
		Type:     xdb.MySQL,
		Database: "my_test_db",
		LogLevel: xdb.Error,
	}
	db, mock, err := xdb.NewDBMock(c)
	require.NoError(t, err)

	mock.ExpectPrepare("^SELECT 1").ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"ok"}).AddRow("1"))
	mock.ExpectPrepare("^SELECT VERSION()").ExpectQuery().WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow("5.7.31"))

	checker := NewChecker(c.Type, db)
	assert.NotNil(t, checker)

	h := checker.Check(context.Background())
	assert.True(t, h.IsUp())

	b, err := json.Marshal(h)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"UP","version":"5.7.31"}`, string(b))
}
