package admin_driver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gitee.com/fatzeng/switch-sdk-core/logger"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

func NewGormLogger() gormlogger.Interface {
	return &gormLoggerAdapter{
		logLevel:                  gormlogger.Info,
		slowThreshold:             200 * time.Millisecond,
		ignoreRecordNotFoundError: true,
	}
}

type gormLoggerAdapter struct {
	logLevel                  gormlogger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
	parameterizedQueries      bool
	colorful                  bool
}

func (l *gormLoggerAdapter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

func (l *gormLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Info {
		logger.Logger.Infof(msg, data...)
	}
}

func (l *gormLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Warn {
		logger.Logger.Warnf(msg, data...)
	}
}

func (l *gormLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Error {
		logger.Logger.Errorf(msg, data...)
	}
}

func (l *gormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	latency := float64(elapsed.Nanoseconds()) / 1e6
	sql, rows := fc()

	if l.parameterizedQueries {
		sql = "(SQL query hidden for security)"
	}

	logFields := fmt.Sprintf("[%.3fms] [rows:%v] %s", latency, rows, sql)

	if err != nil && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.ignoreRecordNotFoundError) {
		if l.colorful {
			logger.Logger.Errorf("%s%s - GORM ERROR: %v%s", colorRed, logFields, err, colorReset)
		} else {
			logger.Logger.Errorf("%s - GORM ERROR: %v", logFields, err)
		}
	} else if elapsed > l.slowThreshold && l.slowThreshold != 0 && l.logLevel >= gormlogger.Warn {
		if l.colorful {
			logger.Logger.Warnf("%s%s - SLOW SQL%s", colorYellow, logFields, colorReset)
		} else {
			logger.Logger.Warnf("%s - SLOW SQL", logFields)
		}
	} else if l.logLevel >= gormlogger.Info {
		if l.colorful {
			logger.Logger.Infof("%s%s%s", colorGreen, logFields, colorReset)
		} else {
			logger.Logger.Infof(logFields)
		}
	}
}
