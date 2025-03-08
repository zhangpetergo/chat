// Package logger 用来创建 logger
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

// Logger 及其需要性能的时候使用 logger，logger只支持结构化记录
var Logger *zap.Logger

// Log SugaredLogger 一般不需要性能的场景中使用 SugaredLogger
var Log *zap.SugaredLogger

func InitLogger() {

	// 高于 error level 的进入 error.log 文件
	highLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapcore.ErrorLevel
	})

	// 小于 error level 并且大于 debug level 的进入 info.log 文件
	lowLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level < zapcore.ErrorLevel && level > zapcore.DebugLevel
	})

	// 创建 WriteSyncer
	infoFileWriteSyncer := GetInfoWriteSyncer("")
	errorWriteSyncer := GetErrorWriteSyncer("")

	infoCore := zapcore.NewCore(getEncoder(), zapcore.NewMultiWriteSyncer(infoFileWriteSyncer, os.Stdout), lowLevel)
	errorCore := zapcore.NewCore(getEncoder(), zapcore.NewMultiWriteSyncer(errorWriteSyncer, os.Stdout), highLevel)

	var coreArr []zapcore.Core

	coreArr = append(coreArr, infoCore)
	coreArr = append(coreArr, errorCore)
	//coreArr = append(coreArr, consoleCore)

	logger := zap.New(zapcore.NewTee(coreArr...), zap.AddCaller()) // AddCaller() 显示文件名和行号

	Log = logger.Sugar()

	// 确保日志缓冲区在程序退出时被刷新
	defer Log.Sync()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()

	//TimeKey:        "ts",
	//LevelKey:       "level",
	//NameKey:        "logger",
	//CallerKey:      "caller",
	//FunctionKey:    zapcore.OmitKey,
	//MessageKey:     "msg",
	//StacktraceKey:  "stacktrace",
	//LineEnding:     zapcore.DefaultLineEnding,
	//EncodeLevel:    zapcore.LowercaseLevelEncoder,
	//EncodeTime:     zapcore.EpochTimeEncoder,
	//EncodeDuration: zapcore.SecondsDurationEncoder,
	//EncodeCaller:   zapcore.ShortCallerEncoder,

	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 在日志中使用大写字母记录日志级别
	// 比如 INFO
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 不同级别的日志显示不同的颜色
	//encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	return zapcore.NewConsoleEncoder(encoderConfig)

}

func getWriterSyncer(fileName string) zapcore.WriteSyncer {

	lumberWriteSyncer := &lumberjack.Logger{
		Filename:   "./logs" + fileName,
		MaxSize:    10, // megabytes
		MaxBackups: 100,
		MaxAge:     28,    // days
		Compress:   false, //Compress确定是否应该使用gzip压缩已旋转的日志文件。默认值是不执行压缩。
	}

	return zapcore.AddSync(lumberWriteSyncer)
}

func GetMultiWriteSyncer(fileName string) zapcore.WriteSyncer {
	return zapcore.NewMultiWriteSyncer(getWriterSyncer(fileName), os.Stdout)
}

func GetInfoWriteSyncer(fileName string) zapcore.WriteSyncer {
	lumberWriteSyncer := &lumberjack.Logger{
		Filename:   "./logs/info.log",
		MaxSize:    10, // megabytes
		MaxBackups: 100,
		MaxAge:     28,    // days
		Compress:   false, //Compress确定是否应该使用gzip压缩已旋转的日志文件。默认值是不执行压缩。
	}

	return zapcore.AddSync(lumberWriteSyncer)
}

func GetErrorWriteSyncer(fileName string) zapcore.WriteSyncer {
	lumberWriteSyncer := &lumberjack.Logger{
		Filename:   "./logs/error.log",
		MaxSize:    10, // megabytes
		MaxBackups: 100,
		MaxAge:     28,    // days
		Compress:   false, //Compress确定是否应该使用gzip压缩已旋转的日志文件。默认值是不执行压缩。
	}

	return zapcore.AddSync(lumberWriteSyncer)
}
