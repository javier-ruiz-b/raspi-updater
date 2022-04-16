package nlog

import (
	"container/list"
	"fmt"
	"log"
	"time"
)

var logEntries = list.New()

type LogEntry struct {
	timestamp time.Time
	logType   LogType
	text      string
}

func Debug(items ...any) {
	Log(LogDebug, items...)
}

func Error(items ...any) {
	Log(LogError, items...)
}

func Log(logType LogType, items ...any) {
	result := LogEntry{
		timestamp: time.Now(),
		logType:   logType,
		text:      fmt.Sprint(items...),
	}

	logEntries.PushBack(result)

	log.Println(result.text)
}

type LogType byte

const (
	LogDebug   LogType = 0
	LogInfo            = 1
	LogWarning         = 2
	LogError           = 3
	LogPanic           = 4
)
