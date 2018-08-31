package logger

import (
	"github.com/devfeel/dotweb/framework/file"
	"sync"
)

const (
	// LogLevelDebug debug log level
	LogLevelDebug = "DEBUG"
	// LogLevelInfo info log level
	LogLevelInfo  = "INFO"
	// LogLevelWarn warn log level
	LogLevelWarn  = "WARN"
	// LogLevelError error log level
	LogLevelError = "ERROR"
)

const(
	LogTarget_InnerLog = "InnerLog"
	LogTarget_LoadBalance = "LoadBalance"
	LogTarget_Gateway = "Gateway"
	LogTarget_JsonRpc = "JsonRpc"
	LogTarget_Default  = "Default"
	LogTarget_UIServer = "UIServer"
)

var (
	GatewayLogger Logger
	JsonRpcLogger Logger
	InnerLogger Logger
	LoadBalanceLogger Logger
	DefaultLogger Logger
	UIServerLogger Logger
)

type Logger interface {
	SetLogPath(logPath string)
	SetEnabledConsole(enabled bool)
	SetEnabledLog(enabledLog bool)
	Debug(log string)
	Print(log string)
	Info(log string)
	Warn(log string)
	Error(log string)
}

var (
	loggerMap map[string]Logger
	loggerMutex *sync.RWMutex
	DefaultLogPath string
	EnabledLog     bool = false
	EnabledConsole bool = false
)

func init(){
	loggerMap = make(map[string]Logger)
	loggerMutex = new(sync.RWMutex)

	DefaultLogPath = file.GetCurrentDirectory() +"/logs"

	GatewayLogger = GetLogger(LogTarget_Gateway)
	InnerLogger = GetLogger(LogTarget_InnerLog)
	LoadBalanceLogger = GetLogger(LogTarget_LoadBalance)
	DefaultLogger = GetLogger(LogTarget_Default)
	JsonRpcLogger = GetLogger(LogTarget_JsonRpc)
	UIServerLogger = GetLogger(LogTarget_UIServer)
}

// GetLogger get logger with log target
func GetLogger(target string) Logger {
	loggerMutex.RLock()
	logger, exists := loggerMap[target]
	loggerMutex.RUnlock()
	if !exists{
		loggerMutex.Lock()
		logger = NewXLog(target)
		loggerMap[target] = logger
		loggerMutex.Unlock()
	}
	return logger
}


