package xdb

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/conf"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/sliveryou/go-tool/v2/timex"
)

// Config 数据库相关配置
type Config struct {
	Type            Type          `json:",default=mysql,options=[mysql,postgres,sqlite,sqlserver]"` // 数据库类型，枚举（mysql、postgres、sqlite 和 sqlserver）
	Host            string        `json:",optional"`                                                // 地址
	Port            int           `json:",optional"`                                                // 端口
	User            string        `json:",optional"`                                                // 用户
	Password        string        `json:",optional"`                                                // 密码
	Database        string        // 数据库（数据库类型为 sqlite 时，为 db 文件地址）
	Params          string        `json:",optional"`                                      // 额外 DSN 参数
	MaxIdleConns    int           `json:",default=10"`                                    // 最大空闲连接数
	MaxOpenConns    int           `json:",default=50"`                                    // 最大打开连接数
	ConnMaxLifeTime time.Duration `json:",default=1h"`                                    // 连接最大生存时间
	ConnMaxIdleTime time.Duration `json:",default=1h"`                                    // 连接最大空闲时间
	LogLevel        LogLevel      `json:",default=info,options=[info,warn,error,silent]"` // 日志级别，枚举（info、warn、error 和 silent）
	SlowThreshold   time.Duration `json:",default=200ms"`                                 // 慢查询阈值
}

// NewDB 新建 gorm.DB  对象
func NewDB(c Config) (*gorm.DB, error) {
	err := c.fillDefault()
	if err != nil {
		return nil, errors.WithMessage(err, "xdb: fill default db config err")
	}

	db, err := c.open()
	if err != nil {
		return nil, errors.WithMessage(err, "xdb: open db connection err")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.WithMessage(err, "xdb: get db instance err")
	}

	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(c.ConnMaxLifeTime)
	sqlDB.SetConnMaxIdleTime(c.ConnMaxIdleTime)

	return db, nil
}

// MustNewDB 新建 gorm.DB 对象
func MustNewDB(c Config) *gorm.DB {
	db, err := NewDB(c)
	if err != nil {
		panic(err)
	}

	return db
}

// open 打开数据库连接
func (c Config) open() (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch c.Type {
	case MySQL:
		dialector = mysql.New(c.GetMySQLConfig())
	case PostgreSQL:
		dialector = postgres.New(c.GetPostgreSQLConfig())
	case SQLite:
		dialector = sqlite.Open(c.GetSQLiteDSN())
	case SQLServer:
		dialector = sqlserver.New(c.GetSQLServerConfig())
	default:
		dialector = mysql.New(c.GetMySQLConfig())
	}

	return gorm.Open(dialector, c.GetGORMConfig())
}

// GetGORMConfig 获取 GORM 相关配置
func (c Config) GetGORMConfig() *gorm.Config {
	gc := &gorm.Config{
		PrepareStmt:     true, // 缓存预编译语句
		QueryFields:     true, // 根据字段名称查询
		CreateBatchSize: 100,  // 批次创建大小
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 数据表名单数
		},
		Logger:  NewLogger(c.LogLevel.ToGROMLogLevel(), c.SlowThreshold), // 设置日志记录器
		NowFunc: func() time.Time { return timex.Now() },                 // 当前时间载入时区
	}

	return gc
}

// GetMySQLConfig 获取 GORM MySQL 相关配置
func (c Config) GetMySQLConfig() mysql.Config {
	// https://github.com/go-gorm/mysql
	// https://github.com/go-sql-driver/mysql#dsn-data-source-name
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?", c.User, c.Password, c.Host, c.Port, c.Database)
	if c.Params != "" {
		dsn += c.Params
	} else {
		dsn += "charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
	}

	return mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         255,   // string 类型字段默认长度
		SkipInitializeWithVersion: false, // 禁用根据当前 mysql 版本自动配置
	}
}

// GetPostgreSQLConfig 获取 GORM PostgreSQL 相关配置
func (c Config) GetPostgreSQLConfig() postgres.Config {
	// https://github.com/go-gorm/postgres
	// https://github.com/jackc/pgx
	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s ", c.User, c.Password, c.Host, c.Port, c.Database)
	if c.Params != "" {
		dsn += c.Params
	} else {
		dsn += "sslmode=disable TimeZone=Asia/Shanghai"
	}

	return postgres.Config{
		DSN: dsn,
	}
}

// GetSQLiteDSN 获取 GORM SQLite DSN 配置
func (c Config) GetSQLiteDSN() string {
	// https://github.com/glebarez/sqlite
	// https://github.com/glebarez/go-sqlite
	dsn := fmt.Sprintf("%s?_pragma=foreign_keys(1)&", c.Database)
	if c.Params != "" {
		dsn += c.Params
	} else {
		dsn += "_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)"
	}

	return dsn
}

// GetSQLServerConfig 获取 GORM SQLServer 相关配置
func (c Config) GetSQLServerConfig() sqlserver.Config {
	// https://github.com/go-gorm/sqlserver
	// https://github.com/microsoft/go-mssqldb
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", c.User, c.Password, c.Host, c.Port, c.Database)
	if c.Params != "" {
		dsn += "&" + c.Params
	}

	return sqlserver.Config{
		DSN:               dsn,
		DefaultStringSize: 255,
	}
}

// fillDefault 填充默认值
func (c *Config) fillDefault() error {
	fill := &Config{}
	if err := conf.FillDefault(fill); err != nil {
		return err
	}

	if c.Type == "" {
		c.Type = fill.Type
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = fill.MaxIdleConns
	}
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = fill.MaxOpenConns
	}
	if c.ConnMaxLifeTime == 0 {
		c.ConnMaxLifeTime = fill.ConnMaxLifeTime
	}
	if c.ConnMaxIdleTime == 0 {
		c.ConnMaxIdleTime = fill.ConnMaxIdleTime
	}
	if c.LogLevel == "" {
		c.LogLevel = fill.LogLevel
	}
	if c.SlowThreshold == 0 {
		c.SlowThreshold = fill.SlowThreshold
	}

	return nil
}
