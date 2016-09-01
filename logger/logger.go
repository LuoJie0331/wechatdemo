package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"gitlab.appdao.com/luojie/wechat/conf"
	"gitlab.appdao.com/luojie/wechat/utils"
)

var (
	nowDate    string
	rotateLock sync.RWMutex

	CommonLogger  *Logger
	RequestLogger *Logger
	ChargeLogger  *Logger
	AutoLogger    *Logger
)

const (
	LOG_LEVEL_DEBUG = logrus.DebugLevel
	LOG_LEVEL_INFO  = logrus.InfoLevel
	LOG_LEVEL_WARN  = logrus.WarnLevel
	LOG_LEVEL_ERROR = logrus.ErrorLevel
	LOG_LEVEL_FATAL = logrus.FatalLevel
	LOG_LEVEL_PANIC = logrus.PanicLevel
)

func init() {
	nowDate = utils.GetNowStringYMD()

	if conf.DebugMode {
		CommonLogger = newLogger(conf.ServiceConfig.LogDir, "log", LOG_LEVEL_DEBUG)
		ChargeLogger = newLogger(conf.ServiceConfig.LogDir, "charge", LOG_LEVEL_DEBUG)
		RequestLogger = newLogger(conf.ServiceConfig.LogDir, "request", LOG_LEVEL_DEBUG)
		AutoLogger = newLogger(conf.ServiceConfig.LogDir, "auto", LOG_LEVEL_DEBUG)
	} else {
		CommonLogger = newLogger(conf.ServiceConfig.LogDir, "log", LOG_LEVEL_INFO)
		ChargeLogger = newLogger(conf.ServiceConfig.LogDir, "charge", LOG_LEVEL_INFO)
		RequestLogger = newLogger(conf.ServiceConfig.LogDir, "request", LOG_LEVEL_INFO)
		AutoLogger = newLogger(conf.ServiceConfig.LogDir, "auto", LOG_LEVEL_INFO)
	}

	utils.DoEverTask(rotateLog, "rotate_log", 5*time.Second)
}

type Logger struct {
	*logrus.Logger

	path         string
	fileNameBase string
	file         *os.File
}

func (logger *Logger) init() {
	p := filepath.Join(logger.path, fmt.Sprintf("%s.log.%s", logger.fileNameBase, nowDate))
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
	if err != nil {
		log.Printf("Failed to init [%s] log: %s", logger.fileNameBase, err.Error())
		os.Exit(1)
	}

	logger.file = f
	logger.Out = f
	logger.Formatter = new(logrus.JSONFormatter)
}

func (logger *Logger) rotate() {
	logger.file.Close()
	p := filepath.Join(logger.path, fmt.Sprintf("%s.log.%s", logger.fileNameBase, nowDate))
	f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
	if err != nil {
		logger.Fatalf("Failed to init [%s] log: %s", logger.fileNameBase, err.Error())
	}

	logger.Out = f
	logger.Formatter = new(logrus.JSONFormatter)
}

func (logger *Logger) newLog(level logrus.Level, data map[string]interface{}) {
	rotateLock.RLock()
	defer rotateLock.RUnlock()

	fields := logrus.Fields(data)
	switch level {
	case LOG_LEVEL_DEBUG:
		logger.WithFields(fields).Debug("")
	case LOG_LEVEL_INFO:
		logger.WithFields(fields).Info("")
	case LOG_LEVEL_WARN:
		logger.WithFields(fields).Warn("")
	case LOG_LEVEL_ERROR:
		logger.WithFields(fields).Error("")
	case LOG_LEVEL_FATAL:
		logger.WithFields(fields).Fatal("")
	case LOG_LEVEL_PANIC:
		logger.WithFields(fields).Panic("")
	default:
		logger.WithFields(fields).Info("")
	}
}

func (logger *Logger) Debug(data map[string]interface{}) {
	logger.newLog(LOG_LEVEL_DEBUG, data)
}

func (logger *Logger) Warn(data map[string]interface{}) {
	logger.newLog(LOG_LEVEL_WARN, data)
}

func (logger *Logger) Info(data map[string]interface{}) {
	logger.newLog(LOG_LEVEL_INFO, data)
}

func (logger *Logger) Error(data map[string]interface{}) {
	logger.newLog(LOG_LEVEL_ERROR, data)
}

func (logger *Logger) Fatal(data map[string]interface{}) {
	logger.newLog(LOG_LEVEL_FATAL, data)
}

func (logger *Logger) Panic(data map[string]interface{}) {
	logger.newLog(LOG_LEVEL_PANIC, data)
}

func (logger *Logger) SetLevel(level logrus.Level) {
	logger.Level = level
}

func newLogger(path string, baseName string, level logrus.Level) *Logger {
	l := &Logger{new(logrus.Logger), path, baseName, nil}
	l.init()
	l.SetLevel(level)

	return l
}

func rotateLog() time.Duration {
	_nowDate := utils.GetNowStringYMD()
	if _nowDate == nowDate {
		now := time.Now()
		return time.Duration(23-now.Hour())*time.Hour + time.Duration(59-now.Minute())*time.Minute + time.Duration(60-now.Second())*time.Second
	}

	rotateLock.Lock()
	defer rotateLock.Unlock()

	nowDate = _nowDate
	CommonLogger.rotate()
	ChargeLogger.rotate()
	AutoLogger.rotate()
	RequestLogger.rotate()
	now := time.Now()
	return time.Duration(23-now.Hour())*time.Hour + time.Duration(59-now.Minute())*time.Minute + time.Duration(60-now.Second())*time.Second
}
