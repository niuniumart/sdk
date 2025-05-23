package nlog

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
	log2 "gorm.io/gorm/logger"
)

type GormLogger struct {
	slowThreshold time.Duration // 慢查询SQL的阈值
}

// NewGormLogger 传参单位为ms
func NewGormLogger(slowThresholdMillisecond int64) *GormLogger {
	return &GormLogger{
		slowThreshold: time.Duration(slowThresholdMillisecond) * time.Millisecond,
	}
}

func (l *GormLogger) LogMode(level log2.LogLevel) log2.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	Infof(ctx, s, i...)
}

func (l *GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	Warnf(ctx, s, i...)
}

func (l *GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	Errorf(ctx, s, i...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {

	// 获取运行时间
	executeTime := time.Since(begin)
	// 获取 SQL 请求和返回条数
	sql, rows := fc()

	// Gorm 错误
	if err != nil {
		// 记录未找到的错误使用 warning 等级
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Infof(ctx, "Database ErrRecordNotFound，sql: %s, time: %s, rows: %d", sql, executeTime.String(), rows)
		} else {
			// 其他错误使用 error 等级
			Errorf(ctx, "Database Error sql: %s, time: %s, rows: %d, err: %v", sql, executeTime.String(), rows, err)
		}
		return
	}

	// 慢查询日志
	if l.slowThreshold != 0 && executeTime > l.slowThreshold {
		Infof(ctx, "Database Slow Log sql: %s, time: %s, rows: %d", sql, executeTime.String(), rows)
	} else {
		// 记录所有 SQL 请求
		Infof(ctx, "Database Query: %s, time: %s, rows: %d, err: %v", sql, executeTime.String(), rows, err)

	}

}
