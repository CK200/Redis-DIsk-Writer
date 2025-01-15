package logs

import (
	"fmt"
	"log"
	"main/pkg/globals"
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var QueueLog *zap.Logger

func SetUpQueueLogs() {
	config := zap.NewProductionConfig()

	// Set log file path
	logFilePath := globals.ApplicationConfig.Application.LogPath + "queuesBackUp.log"
	config.OutputPaths = []string{logFilePath}

	// Create lumberjack logger for log rotation
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    5,  // MB
		MaxAge:     28, // Days
		MaxBackups: 30000,
		LocalTime:  true,
		Compress:   true,
	}

	// Add lumberjack logger as output
	// config.OutputPaths = append(config.OutputPaths, "stdout")           // Also log to stdout
	config.ErrorOutputPaths = append(config.ErrorOutputPaths, "stderr") // Log errors to stderr
	config.EncoderConfig.TimeKey = ""
	config.EncoderConfig.LevelKey = ""

	// config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // Use ISO8601 time format
	// config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	// config.Callers = zapcore.ShortCaller

	// Create Zap core with log rotation
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(config.EncoderConfig),
		zapcore.AddSync(lumberjackLogger),
		zap.NewAtomicLevelAt(zap.InfoLevel), // Change log level here if needed
		// zap.RegisterEncoder("",zapcore.ShortCallerEncoder())
	)

	// Create logger with the configured core
	logger := zap.New(core)

	QueueLog = logger

	// Close the lumberjack logger when application exits
	defer lumberjackLogger.Close()

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			fileInfo, err := os.Stat(logFilePath)
			if err != nil {
				log.Println("Failed to get log file info :: ", err.Error())
				continue
			}

			// Check if file size is greater than 0 or exceeds 5 MB
			if fileInfo.Size() > 0 {
				err := lumberjackLogger.Rotate()
				if err != nil {
					log.Println("Failed to rotate log file", err.Error())
				}
			}
		}
	}()
}

func InfoLog(format string, a ...interface{}) {
	stringMessage := fmt.Sprintf(format, a...)
	QueueLog.Info(stringMessage)
}

func ErrorLog(format string, a ...interface{}) {
	stringMessage := fmt.Sprintf(format, a...)
	QueueLog.Error(stringMessage)
}
