package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitee.com/fatzeng/switch-sdk-core/logger"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// ZapLogger 封装了 zap.Logger 同时实现了golang原生的logger
type ZapLogger struct {
	*zap.Logger
	loggerCfg *logger.LoggerConfig
	sugar     *zap.SugaredLogger
}

// New 创建一个新的日志记录器
func New(loggerCfg *logger.LoggerConfig) (*ZapLogger, error) {
	if loggerCfg == nil {
		loggerCfg = logger.DefaultLogConfig()
	} else {
		loggerCfg.Initial()
	}
	if err := os.MkdirAll(loggerCfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("create log dir fail: %v", err)
	}
	level, err := zap.ParseAtomicLevel(loggerCfg.Level)
	if err != nil {
		return nil, fmt.Errorf("load log level fail: %v", err)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        loggerCfg.LogConfigMeta.TimeKey,
		LevelKey:       loggerCfg.LogConfigMeta.LevelKey,
		NameKey:        loggerCfg.LogConfigMeta.NameKey,
		CallerKey:      loggerCfg.LogConfigMeta.CallerKey,
		MessageKey:     loggerCfg.LogConfigMeta.MessageKey,
		StacktraceKey:  loggerCfg.LogConfigMeta.StacktraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(loggerCfg.TimeFormat),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var cores []zapcore.Core
	goFileNameFormat := convertDateFormat(loggerCfg.FileNameFormat)
	filename := time.Now().Format(goFileNameFormat)
	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(loggerCfg.OutputDir, filename),
		MaxSize:    loggerCfg.MaxSize,
		MaxBackups: loggerCfg.MaxBackups,
		MaxAge:     loggerCfg.MaxAge,
		Compress:   loggerCfg.Compress,
	}

	var encoder zapcore.Encoder
	if loggerCfg.EnableJSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	cores = append(cores, zapcore.NewCore(
		encoder,
		zapcore.AddSync(fileWriter),
		level,
	))

	// 控制台输出 非线上环境最好不要打开
	if loggerCfg.EnableConsole {
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			level,
		))
	}

	core := zapcore.NewTee(cores...)
	options := make([]zap.Option, 0)
	if loggerCfg.EnableStackTrace {
		trace, err := zapcore.ParseLevel(loggerCfg.StackTraceLevel)
		if err != nil {
			return nil, fmt.Errorf("load log stack trace level fail: %v", err)
		}
		options = append(options, zap.AddStacktrace(trace))
	}

	if loggerCfg.ShowCaller {
		options = append(options, zap.AddCaller())
	}

	// 添加业务自定义的原字段
	if len(loggerCfg.CustomFields) > 0 {
		fields := make([]zap.Field, 0, len(loggerCfg.CustomFields))
		for k, v := range loggerCfg.CustomFields {
			fields = append(fields, zap.Any(k, v))
		}
		options = append(options, zap.Fields(fields...))
	}

	zapLogger := zap.New(core, options...)

	return &ZapLogger{
		Logger:    zapLogger,
		loggerCfg: loggerCfg,
		sugar:     zapLogger.Sugar(),
	}, nil
}

// FieldF 创建一个日志字段
func FieldF(key string, value interface{}) zapcore.Field {
	switch v := value.(type) {
	case string:
		return zap.String(key, v)
	case int:
		return zap.Int(key, v)
	case int64:
		return zap.Int64(key, v)
	case float64:
		return zap.Float64(key, v)
	case bool:
		return zap.Bool(key, v)
	case time.Time:
		return zap.Time(key, v)
	case time.Duration:
		return zap.Duration(key, v)
	case error:
		return zap.Error(v)
	default:
		return zap.Any(key, v)
	}
}

func (l *ZapLogger) Sync() error {
	return l.Logger.Sync()
}

func (l *ZapLogger) with(fields ...zapcore.Field) *ZapLogger {
	return &ZapLogger{
		Logger:    l.Logger.With(fields...),
		loggerCfg: l.loggerCfg,
		sugar:     l.Logger.With(fields...).Sugar(),
	}
}

func (l *ZapLogger) Sugar() *zap.SugaredLogger {
	return l.sugar
}

// debug
func (l *ZapLogger) Debug(args ...interface{}) {
	l.sugar.Debug(args...)
}
func (l *ZapLogger) Debugf(template string, args ...interface{}) {
	l.sugar.Debugf(template, args...)
}

// info
func (l *ZapLogger) Info(args ...interface{}) {
	l.sugar.Info(args...)
}
func (l *ZapLogger) Infof(template string, args ...interface{}) {
	l.sugar.Infof(template, args...)
}

// Warn
func (l *ZapLogger) Warn(args ...interface{}) {
	l.sugar.Warn(args...)
}
func (l *ZapLogger) Warnf(template string, args ...interface{}) {
	l.sugar.Warnf(template, args...)
}

// Error
func (l *ZapLogger) Error(args ...interface{}) {
	l.sugar.Error(args...)
}
func (l *ZapLogger) Errorf(template string, args ...interface{}) {
	l.sugar.Errorf(template, args...)
}

// Fatal
func (l *ZapLogger) Fatal(args ...interface{}) {
	l.sugar.Fatal(args...)
}
func (l *ZapLogger) Fatalf(template string, args ...interface{}) {
	l.sugar.Fatalf(template, args...)
}

// Panic
func (l *ZapLogger) Panic(args ...interface{}) {
	l.sugar.Panic(args...)
}
func (l *ZapLogger) Panicf(template string, args ...interface{}) {
	l.sugar.Panicf(template, args...)
}

// Print implements standard logging Print
func (l *ZapLogger) Print(v ...interface{}) {
	l.Info(v...)
}

// Printf implements standard logging Printf
func (l *ZapLogger) Printf(format string, v ...interface{}) {
	l.Infof(format, v...)
}

// Println implements standard logging Println
func (l *ZapLogger) Println(v ...interface{}) {
	l.Info(v...)
}

// Fatalln implements standard logging Fatalln
func (l *ZapLogger) Fatalln(v ...interface{}) {
	l.Fatal(v...)
}

// Panicln implements standard logging Panicln
func (l *ZapLogger) Panicln(v ...interface{}) {
	l.Panic(v...)
}

func (l *ZapLogger) With(fields map[string]interface{}) logger.ILogger {
	zapFields := make([]zapcore.Field, 0, len(fields))
	for key, keyVal := range fields {
		zapFields = append(zapFields, FieldF(key, keyVal))
	}
	return l.with(zapFields...)
}

// 这里对年月日等常用标识转换成go中的标识
func convertDateFormat(format string) string {
	replacements := map[string]string{
		"%Y": "2006",
		"%y": "06",
		"%m": "01",
		"%d": "02",
		"%H": "15",
		"%I": "03",
		"%M": "04",
		"%S": "05",
		"%p": "PM",
		"%Z": "MST",
	}

	result := format
	for placeholder, goFormat := range replacements {
		result = strings.Replace(result, placeholder, goFormat, -1)
	}
	return result
}
