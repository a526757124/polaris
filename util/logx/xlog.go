package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"github.com/devfeel/polaris/util/filex"
)

type chanLog struct {
	Content   string
	LogTarget string
	LogLevel  string
	isRaw     bool
}

type xLog struct {
	logRootPath    string
	logTarget string
	logChan chan chanLog
	enabledLog     bool
	enabledConsole bool
}

//create new xLog
func NewXLog(target string) *xLog {
	l := &xLog{logTarget:target, logChan: make(chan chanLog, 10000)}
	go l.handleCustom()
	return l
}

const (
	defaultDateFormatForFileName = "2006_01_02"
	defaultDateLayout            = "2006-01-02"
	defaultFullTimeLayout        = "2006-01-02 15:04:05.9999"
	defaultTimeLayout            = "2006-01-02 15:04:05"
)

// Debug debug log with default format
func (l *xLog) Debug(log string) {
	l.log(log, LogLevelDebug, false)
}

// Print debug log with no format
func (l *xLog) Print(log string) {
	l.log(log, LogLevelDebug, true)
}

// Info info log with default format
func (l *xLog) Info(log string) {
	l.log(log, LogLevelInfo, false)
}

// Warn warn log with default format
func (l *xLog) Warn(log string) {
	l.log(log, LogLevelWarn, false)
}

// Error error log with default format
func (l *xLog) Error(log string) {
	l.log(log, LogLevelError, false)
}

// log push log into chan
func (l *xLog) log(log string, logLevel string, isRaw bool) {
	if l.enabledLog {
		chanLog := chanLog{
			LogTarget: l.logTarget + "_" + logLevel,
			Content:   log,
			LogLevel:  logLevel,
			isRaw:     isRaw,
		}
		l.logChan <- chanLog
	}
}

//SetLogPath set log path
func (l *xLog) SetLogPath(rootPath string) {
	//设置日志根目录
	l.logRootPath = rootPath
	if !strings.HasSuffix(l.logRootPath, "/") {
		l.logRootPath = l.logRootPath + "/"
	}
}

//SetEnabledLog set enabled log
func (l *xLog) SetEnabledLog(enabledLog bool) {
	l.enabledLog = enabledLog
}

//SetEnabledConsole set enabled Console output
func (l *xLog) SetEnabledConsole(enabled bool) {
	l.enabledConsole = enabled
}

//处理日志内部函数
func (l *xLog) handleCustom() {
	for {
		log := <-l.logChan
		l.writeLog(log, "custom")
	}
}

func (l *xLog) writeLog(chanLog chanLog, level string) {
	filePath := l.logRootPath + chanLog.LogTarget
	switch level {
	case "custom":
		filePath = filePath + "_" + time.Now().Format(defaultDateFormatForFileName) + ".log"
		break
	}
	log := chanLog.Content
	if !chanLog.isRaw {
		log = fmt.Sprintf("%s [%s] %s", time.Now().Format(defaultFullTimeLayout), chanLog.LogLevel, chanLog.Content)
	}
	if l.enabledConsole {
		fmt.Println(log)
	}
	writeFile(filePath, log)
}

func writeFile(logFile string, log string) {
	pathDir := filepath.Dir(logFile)
	if !filex.Exist(pathDir) {
		//create path
		err := os.MkdirAll(pathDir, 0777)
		if err != nil {
			fmt.Println("xlog.writeFile create path error ", err)
			return
		}
	}

	var mode os.FileMode
	flag := syscall.O_RDWR | syscall.O_APPEND | syscall.O_CREAT
	mode = 0666
	logstr := log + "\r\n"
	file, err := os.OpenFile(logFile, flag, mode)
	defer file.Close()
	if err != nil {
		fmt.Println(logFile, err)
		return
	}
	file.WriteString(logstr)
}
