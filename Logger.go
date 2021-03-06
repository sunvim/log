package log

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ssgo/standard"
)

const LogFilePath = "file" //日志输出位置 文件:行号

type LevelType int

const DEBUG LevelType = 1
const INFO LevelType = 2
const WARNING LevelType = 3
const ERROR LevelType = 4

var globalLevelType LevelType = INFO

type Logger struct {
	level       LevelType
	truncations []string
	writer      func(string)
}

func (l *Logger) SetGlobalLevel(level LevelType) {
	globalLevelType = level
}

func (logger *Logger) SetLevel(level LevelType) {
	logger.level = level
}

func (logger *Logger) SetWriter(writer func(string)) {
	logger.writer = writer
}

func (logger *Logger) SetTruncations(truncations ...string) {
	logger.truncations = append(logger.truncations, truncations...)
}

func (logger *Logger) Debug(logType string, data ...interface{}) {
	output(logger, DEBUG, logType, buildLogData(data...))
}

func (logger *Logger) Info(logType string, data ...interface{}) {
	output(logger, INFO, logType, buildLogData(data...))
}

func (logger *Logger) Warning(logType string, data ...interface{}) {
	logger.trace(WARNING, logType, buildLogData(data...))
}

func (logger *Logger) Error(logType string, data ...interface{}) {
	logger.trace(ERROR, logType, buildLogData(data...))
}

func (logger *Logger) log(logLevel LevelType, logType string, data map[string]interface{}) {
	output(logger, logLevel, logType, data)
}

func output(logger *Logger, logLevel LevelType, logType string, data map[string]interface{}) {
	settedLevel := logger.level
	if settedLevel == 0 {
		settedLevel = globalLevelType
	}
	if logLevel < settedLevel {
		return
	}

	LogLevelName := standard.LogLevelInfo
	switch logLevel {
	case DEBUG:
		LogLevelName = standard.LogLevelDebug
	case INFO:
		LogLevelName = standard.LogLevelInfo
	case WARNING:
		LogLevelName = standard.LogLevelWarning
	case ERROR:
		LogLevelName = standard.LogLevelError
	}

	_, fileName, lineNum, _ := runtime.Caller(2)
	fileName = filepath.Base(fileName)
	b := strings.Builder{}
	b.WriteString(fileName)
	b.WriteString(":")
	b.WriteString(strconv.Itoa(lineNum))
	data[LogFilePath] = b.String()
	data[standard.LogFieldLevel] = LogLevelName
	data[standard.LogFieldTime] = MakeLogTime(time.Now())
	data[standard.LogFieldType] = logType
	buf, err := json.Marshal(data)

	if err != nil {
		// 无法序列化的数据包装为 undefined
		buf, err = json.Marshal(map[string]interface{}{
			LogFilePath:            data[LogFilePath],
			standard.LogFieldLevel: data[standard.LogFieldLevel],
			standard.LogFieldTime:  data[standard.LogFieldTime],
			standard.LogFieldType:  standard.LogTypeUndefined,
			"info":                 fmt.Sprint(data),
		})
		return
	}

	if err == nil {
		if logger.writer == nil {
			log.Print(string(buf))
		} else {
			logger.writer(string(buf))
		}
	}
}

func (logger *Logger) trace(LogLevel LevelType, logType string, data map[string]interface{}) {
	traces := make([]string, 0)
	for i := 1; i < 20; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if strings.Contains(file, "/go/src/") {
			continue
		}
		if strings.Contains(file, "/ssgo/log") {
			continue
		}
		if logger.truncations != nil {
			for _, truncation := range logger.truncations {
				pos := strings.Index(file, truncation)
				if pos != -1 {
					file = file[pos+len(truncation):]
				}
			}
		}
		traces = append(traces, fmt.Sprintf("%s:%d", file, line))
	}
	data[standard.LogFieldTraces] = strings.Join(traces, "; ")
	output(logger, LogLevel, logType, data)
}

func buildLogData(args ...interface{}) map[string]interface{} {
	if len(args) == 1 {
		if mapData, ok := args[0].(map[string]interface{}); ok {
			return mapData
		}
	}
	data := map[string]interface{}{}
	for i := 1; i < len(args); i += 2 {
		if k, ok := args[i-1].(string); ok {
			data[k] = args[i]
		}
	}
	return data
}

func (logger *Logger) LogRequest(app, node, clientIp, fromApp, fromNode, clientId, sessionId, requestId, host, scheme, proto string, authLevel, priority int, method, path string, requestHeaders map[string]string, requestData map[string]interface{}, usedTime float32, responseCode int, responseHeaders map[string]string, responseDataLength uint, responseData interface{}, extraInfo map[string]interface{}) {
	if extraInfo == nil {
		extraInfo = map[string]interface{}{}
	}
	extraInfo[standard.LogFieldRequestApp] = app
	extraInfo[standard.LogFieldRequestNode] = node
	extraInfo[standard.LogFieldRequestClientIp] = clientIp
	extraInfo[standard.LogFieldRequestFromApp] = fromApp
	extraInfo[standard.LogFieldRequestFromNode] = fromNode
	extraInfo[standard.LogFieldRequestClientId] = clientId
	extraInfo[standard.LogFieldRequestSessionId] = sessionId
	extraInfo[standard.LogFieldRequestRequestId] = requestId
	extraInfo[standard.LogFieldRequestHost] = host
	extraInfo[standard.LogFieldRequestScheme] = scheme
	extraInfo[standard.LogFieldRequestProto] = proto
	extraInfo[standard.LogFieldRequestAuthLevel] = authLevel
	extraInfo[standard.LogFieldRequestPriority] = priority
	extraInfo[standard.LogFieldRequestMethod] = method
	extraInfo[standard.LogFieldRequestPath] = path
	extraInfo[standard.LogFieldRequestRequestHeaders] = requestHeaders
	extraInfo[standard.LogFieldRequestArgs] = requestData
	extraInfo[standard.LogFieldRequestUsedTime] = usedTime
	extraInfo[standard.LogFieldRequestStatus] = responseCode
	extraInfo[standard.LogFieldRequestResponseHeaders] = responseHeaders
	extraInfo[standard.LogFieldRequestOutLen] = responseDataLength
	extraInfo[standard.LogFieldRequestResult] = responseData
	logger.log(INFO, standard.LogTypeRequest, extraInfo)

}
