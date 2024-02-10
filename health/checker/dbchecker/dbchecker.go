package dbchecker

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/sliveryou/micro-pkg/health"
	"github.com/sliveryou/micro-pkg/xdb"
)

// Checker 数据库检查器结构详情
type Checker struct {
	t  xdb.Type
	db *gorm.DB
}

// NewChecker 新建数据库检查器
func NewChecker(t xdb.Type, db *gorm.DB) *Checker {
	if db == nil {
		panic(errors.New("nil db is invalid"))
	}

	return &Checker{t: t, db: db}
}

// Check 检查数据库健康情况
func (c *Checker) Check(ctx context.Context) health.Health {
	h := health.NewHealth()
	db := c.db.WithContext(ctx)

	err := c.t.CheckDB(db)
	if err != nil {
		h.Down().AddInfo("error", err.Error())
		return h
	}

	version, err := c.t.VersionDB(db)
	if err != nil {
		h.Down().AddInfo("error", err.Error())
		return h
	}
	h.Up().AddInfo("version", version)

	return h
}
