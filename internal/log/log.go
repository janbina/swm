package log

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var _fileLog *zap.Logger
var fileLog *zap.SugaredLogger

var _consoleLog *zap.Logger
var consoleLog *zap.SugaredLogger

func Init(debugLog string) {
	InitFileLog(debugLog)
	initConsoleLog()
}

func InitFileLog(filePath string) {
	if len(filePath) == 0 {
		return
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    10,
		MaxBackups: 2,
		MaxAge:     0,
		Compress:   false,
	})

	core := zapcore.NewCore(encoder, writer, zapcore.DebugLevel)

	_fileLog = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	fileLog = _fileLog.Sugar()
}

func initConsoleLog() {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("02/Jan/15:04:05"))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	writer := zapcore.AddSync(os.Stderr)

	core := zapcore.NewCore(encoder, writer, zapcore.InfoLevel)

	_consoleLog = zap.New(core)
	consoleLog = _consoleLog.Sugar()
}

func Sync() {
	if _fileLog != nil {
		_ = _fileLog.Sync()
	}
	_ = _consoleLog.Sync()
}

func Debug(template string, args ...interface{}) {
	if fileLog != nil {
		fileLog.Debugf(template, args...)
	}
	consoleLog.Debugf(template, args...)
}

func Info(template string, args ...interface{}) {
	if fileLog != nil {
		fileLog.Infof(template, args...)
	}
	consoleLog.Infof(template, args...)
}

func Warn(template string, args ...interface{}) {
	if fileLog != nil {
		fileLog.Infof(template, args...)
	}
	consoleLog.Infof(template, args...)
}

func Error(template string, args ...interface{}) {
	if fileLog != nil {
		fileLog.Infof(template, args...)
	}
	consoleLog.Infof(template, args...)
}

func Fatal(template string, args ...interface{}) {
	if fileLog != nil {
		fileLog.Fatalf(template, args...)
	}
	consoleLog.Fatalf(template, args...)
}
