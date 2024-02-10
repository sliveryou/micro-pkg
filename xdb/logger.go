package xdb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-stack/stack"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm/logger"
)

const (
	traceInfo  = "%s [%.3fms] [rows:%v] %s"
	traceWarn  = "%s %s [%.3fms] [rows:%v] %s"
	traceError = "%s %s [%.3fms] [rows:%v] %s"
)

// Logger 日志记录器
type Logger struct {
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration
}

// NewLogger 新建日志记录器
func NewLogger(logLevel logger.LogLevel, slowThreshold time.Duration) *Logger {
	return &Logger{LogLevel: logLevel, SlowThreshold: slowThreshold}
}

// LogMode 设置日志记录模式
func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info 日志记录 info
func (l *Logger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Info {
		logx.WithContext(ctx).Infof(msg, data...)
	}
}

// Warn 日志记录 warn
func (l *Logger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Warn {
		logx.WithContext(ctx).Slowf(msg, data...)
	}
}

// Error 日志记录 error
func (l *Logger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= logger.Error {
		logx.WithContext(ctx).Errorf(msg, data...)
	}
}

// Trace 日志记录 trace
func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel > logger.Silent {
		log := logx.WithContext(ctx).WithFields(logx.Field("service", "db"))
		elapsed := time.Since(begin)

		switch {
		case err != nil && l.LogLevel >= logger.Error:
			sql, rows := fc()
			if rows == -1 {
				log.Errorf(traceError, FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				log.Errorf(traceError, FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
			sql, rows := fc()
			slowLog := fmt.Sprintf("Slow SQL Greater Than %v", l.SlowThreshold)
			if rows == -1 {
				log.Slowf(traceWarn, FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				log.Slowf(traceWarn, FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		case l.LogLevel == logger.Info:
			sql, rows := fc()
			if rows == -1 {
				log.Infof(traceInfo, FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				log.Infof(traceInfo, FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		}
	}
}

// FileWithLineNum 获取调用堆栈信息
func FileWithLineNum() string {
	cs := stack.Trace().TrimBelow(stack.Caller(2)).TrimRuntime()

	for _, c := range cs {
		s := fmt.Sprintf("%+v", c)
		if !strings.Contains(s, "gorm.io") &&
			!strings.Contains(s, ".gen.go") {
			return s
		}
	}

	return ""
}
