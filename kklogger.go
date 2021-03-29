package kklogger

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var defaultLoggerPath = "alloc/logs/"
var defaultEnvironment = "default"
var LoggerPath = defaultLoggerPath
var Environment = defaultEnvironment
var AsyncWriteChan = make(chan *AsyncBlob)
var AsyncWrite = true
var Shutdown = false
var HandOver = false
var ReportCaller = true
var loggerHooks []LoggerHook
var asyncStarted = false
var file *os.File
var once sync.Once
var entryMaps = map[string]*logrus.Entry{}
var entryLock = sync.Mutex{}
var level = TraceLevel

type Level logrus.Level

const (
	TraceLevel = Level(logrus.TraceLevel)
	DebugLevel = Level(logrus.DebugLevel)
	InfoLevel  = Level(logrus.InfoLevel)
	WarnLevel  = Level(logrus.WarnLevel)
	ErrorLevel = Level(logrus.ErrorLevel)
)

func init() {
	logrus.SetFormatter(&KKJSONFormatter{})
}

func _Init() {
	once.Do(func() {
		e := os.MkdirAll(LoggerPath, 0755)
		if e == nil {
			logFile, e := os.OpenFile(fmt.Sprintf("%s/current.log", LoggerPath),
				os.O_CREATE|os.O_APPEND|os.O_RDWR, 0755)
			if e != nil {
				println(e.Error())
			}

			file = logFile
		} else {
			println(fmt.Sprintf("can't create logger dir at LoggerPath %s", LoggerPath))
		}

		if e != nil {
			logrus.SetOutput(os.Stdout)
		} else {
			logrus.SetOutput(file)
		}

		logrus.SetReportCaller(ReportCaller)
		if AsyncWrite && !asyncStarted {
			logrus.StandardLogger().SetNoLock()
			asyncStarted = true
			go asyncWriteWorker()
		}

		HandOver = false
	})
}

type AsyncBlob struct {
	logLevel Level
	args     interface{}
}

func Reload() {
	HandOver = true
	if file != nil {
		file.Close()
		once = sync.Once{}
	}
}

func SetLoggerHooks(hooks []LoggerHook) {
	loggerHooks = hooks
}

func _GetLogEntry(severity string) *logrus.Entry {
	entry := entryMaps[fmt.Sprintf("%s_%s", Environment, severity)]
	if entry == nil {
		entryLock.Lock()
		if _, f := entryMaps[fmt.Sprintf("%s_%s", Environment, severity)]; !f {
			entry := logrus.WithFields(logrus.Fields{
				"env":      Environment,
				"severity": severity,
			})

			entry.Logger.SetNoLock()
			entryMaps[fmt.Sprintf("%s_%s", Environment, severity)] = entry
		}

		entry = entryMaps[fmt.Sprintf("%s_%s", Environment, severity)]
		entryLock.Unlock()
	}

	return entry
}

func _Log(logLevel Level, args ...interface{}) {
	_Init()
	if len(args) > 1 {
		args = []interface{}{args}
	}

	if _, ok := args[0].([]interface{}); !ok {
		args[0] = []interface{}{args[0]}
	}

	data := args[0].([]interface{})
	if l := len(data); l == 1 {
		if jsonMsg, ok := data[0].(*JsonMsg); ok {
			if marshal := jsonMsg.Marshal(); marshal != nil {
				args[0] = []interface{}{string(marshal)}
			}
		} else if jsonMsg, ok := data[0].(JsonMsg); ok {
			if marshal := jsonMsg.Marshal(); marshal != nil {
				args[0] = []interface{}{string(marshal)}
			}
		} else if str, ok := data[0].(string); ok {
			jsonMsg := SimpleJsonMsg(str)
			if marshal := jsonMsg.Marshal(); marshal != nil {
				args[0] = []interface{}{string(marshal)}
			}
		}
	} else if l == 2 {
		if str, ok := data[0].(string); ok {
			jsonMsg := &JsonMsg{Type: str, Data: data[1]}
			if marshal := jsonMsg.Marshal(); marshal != nil {
				args[0] = []interface{}{string(marshal)}
			}
		}
	}

	if AsyncWrite {
		AsyncWriteChan <- &AsyncBlob{
			logLevel: logLevel,
			args:     args,
		}
	} else {
		var arg interface{} = args
		_GetLogEntry(_GetSeverity(logLevel)).Log(logrus.Level(logLevel), arg)
		_RunHooks(logLevel, args...)
	}
}

func Trace(args ...interface{}) {
	if !logrus.IsLevelEnabled(logrus.TraceLevel) {
		return
	}

	_Log(TraceLevel, args...)
}

func TraceJ(typeName string, obj interface{}) {
	if !logrus.IsLevelEnabled(logrus.TraceLevel) {
		return
	}

	Trace(NewJsonMsg(typeName, obj))
}

func TraceF(f func() interface{}) {
	if !logrus.IsLevelEnabled(logrus.TraceLevel) {
		return
	}

	go Trace(f())
}

func Debug(args ...interface{}) {
	if !logrus.IsLevelEnabled(logrus.DebugLevel) {
		return
	}

	_Log(DebugLevel, args...)
}

func DebugJ(typeName string, obj interface{}) {
	if !logrus.IsLevelEnabled(logrus.DebugLevel) {
		return
	}

	Debug(NewJsonMsg(typeName, obj))
}

func DebugF(f func() interface{}) {
	if !logrus.IsLevelEnabled(logrus.DebugLevel) {
		return
	}

	go Debug(f())
}

func Info(args ...interface{}) {
	if !logrus.IsLevelEnabled(logrus.InfoLevel) {
		return
	}

	_Log(InfoLevel, args...)
}

func InfoJ(typeName string, obj interface{}) {
	if !logrus.IsLevelEnabled(logrus.InfoLevel) {
		return
	}

	Info(NewJsonMsg(typeName, obj))
}

func InfoF(f func() interface{}) {
	if !logrus.IsLevelEnabled(logrus.InfoLevel) {
		return
	}

	go Info(f())
}

func Warn(args ...interface{}) {
	if !logrus.IsLevelEnabled(logrus.WarnLevel) {
		return
	}

	_Log(WarnLevel, args...)
}

func WarnJ(typeName string, obj interface{}) {
	if !logrus.IsLevelEnabled(logrus.WarnLevel) {
		return
	}

	Warn(NewJsonMsg(typeName, obj))
}

func WarnF(f func() interface{}) {
	if !logrus.IsLevelEnabled(logrus.WarnLevel) {
		return
	}

	go Warn(f())
}

func Error(args ...interface{}) {
	if !logrus.IsLevelEnabled(logrus.ErrorLevel) {
		return
	}

	_Log(ErrorLevel, args...)
}

func ErrorJ(typeName string, obj interface{}) {
	if !logrus.IsLevelEnabled(logrus.ErrorLevel) {
		return
	}

	Error(NewJsonMsg(typeName, obj))
}

func ErrorF(f func() interface{}) {
	if !logrus.IsLevelEnabled(logrus.ErrorLevel) {
		return
	}

	go Error(f())
}

func SetLogLevel(logLevel string) {
	level = TraceLevel
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		level = DebugLevel
	case "INFO":
		level = InfoLevel
	case "WARN":
		level = WarnLevel
	case "ERROR":
		level = ErrorLevel
	}

	logrus.SetLevel(logrus.Level(level))
}

func _RunHooks(logLevel Level, args ...interface{}) {
	defer func() {
		if e := recover(); e != nil {
			if err, ok := e.(error); ok {
				ErrorJ("KKLogger.Hook", err.Error())
			}
		}
	}()

	for _, hook := range loggerHooks {
		switch logLevel {
		case TraceLevel:
			hook.Trace(args...)
		case DebugLevel:
			hook.Debug(args...)
		case InfoLevel:
			hook.Info(args...)
		case WarnLevel:
			hook.Warn(args...)
		case ErrorLevel:
			hook.Error(args...)
		}
	}
}

func GetLogLevel() Level {
	return level
}

func _GetSeverity(logLevel Level) string {
	switch logLevel {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARNING"
	case ErrorLevel:
		return "ERROR"
	default:
		return "DEFAULT"
	}
}

type JsonMsg struct {
	Type string      `json:"type,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

func (j *JsonMsg) Marshal() []byte {
	if jsonMsgBytes, e := json.Marshal(j); e == nil {
		return jsonMsgBytes
	}

	return nil
}

func NewJsonMsg(typeName string, data interface{}) *JsonMsg {
	return &JsonMsg{
		Type: typeName,
		Data: data,
	}
}

func SimpleJsonMsg(data interface{}) *JsonMsg {
	return &JsonMsg{
		Data: data,
	}
}

func asyncWriteWorker() {
	for {
		select {
		case blob := <-AsyncWriteChan:
			if HandOver {
				t := time.NewTicker(time.Second)
				for {
					if !HandOver {
						t.Stop()
						break
					}

					<-t.C
				}
			}

			_GetLogEntry(_GetSeverity(blob.logLevel)).Log(logrus.Level(blob.logLevel), blob.args)
			if cast, ok := blob.args.([]interface{}); ok {
				_RunHooks(blob.logLevel, cast...)
			} else {
				_RunHooks(blob.logLevel, blob.args)
			}
		case <-time.After(time.Millisecond):
			if Shutdown {
				return
			}
		}
	}
}
