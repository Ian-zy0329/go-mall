package dao

import (
	"context"
	"github.com/Ian-zy0329/go-mall/common/logger"
	gormLogger "gorm.io/gorm/logger"
	"time"
)

type GormLogger struct {
	SlowThreshold time.Duration
}

func NewGormLogger() *GormLogger {
	return &GormLogger{
		SlowThreshold: 500 * time.Millisecond,
	}
}

func (l *GormLogger) LogMode(lev gormLogger.LogLevel) gormLogger.Interface {
	return &GormLogger{}
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	logger.New(ctx).Info(msg, "data", data)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	logger.New(ctx).Warn(msg, "data", data)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	logger.New(ctx).Error(msg, "data", data)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	duration := time.Since(begin).Milliseconds()
	sql, rows := fc()
	if err != nil {
		logger.New(ctx).Error("SQL ERROR", "sql", sql, "rows", rows, "dur(ms)", duration)
	}
	if duration > l.SlowThreshold.Milliseconds() {
		logger.New(ctx).Warn("SQL SLOW", "sql", sql, "rows", rows, "dur(ms)", duration)
	} else {
		logger.New(ctx).Debug("SQL", "sql", sql, "rows", rows, "dur(ms)", duration)
	}
}
