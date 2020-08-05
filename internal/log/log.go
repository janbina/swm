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
	initFileLog(debugLog)
	initConsoleLog()
}

func initFileLog(filePath string) {
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

func Infof(template string, args ...interface{}) {
	if fileLog != nil {
		fileLog.Infof(template, args)
	}
	consoleLog.Infof(template, args)
}

func Fatal(args ...interface{}) {
	if fileLog != nil {
		fileLog.Fatal(args)
	}
	consoleLog.Fatal(args)
}

func Fatalf(template string, args ...interface{}) {
	if fileLog != nil {
		fileLog.Fatalf(template, args)
	}
	consoleLog.Fatalf(template, args)
}
