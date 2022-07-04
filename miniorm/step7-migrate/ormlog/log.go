package ormlog

import (
	"io"
	"log"
	"os"
	"sync"
)

var (
	errLog   = log.New(os.Stdout, "\033[31m[ERROR]\033[0m ", log.Ldate|log.Ltime|log.LstdFlags|log.Lshortfile)
	warnLog  = log.New(os.Stdout, "\033[33m[WARN]\033[0m ", log.Ldate|log.Ltime|log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[32m[INFO]\033[0m ", log.Ldate|log.Ltime|log.LstdFlags|log.Lshortfile)
	debugLog = log.New(os.Stdout, "\033[34m[DEBUG]\033[0m ", log.Ldate|log.Ltime|log.LstdFlags|log.Lshortfile)
	loggers  = []*log.Logger{errLog, infoLog}
	mu       sync.Mutex
)

var (
	Error  = errLog.Println
	Errorf = errLog.Printf
	Warn   = warnLog.Println
	Warnf  = warnLog.Printf
	Info   = infoLog.Println
	Infof  = infoLog.Printf
	Debug  = debugLog.Println
	Debugf = debugLog.Printf
)

const (
	DebugLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	Disabled
)

type ormError struct {
	text string
}

func (e *ormError) Error() string {
	return e.text
}

func SetLevel(level int) {
	mu.Lock()
	defer mu.Unlock()

	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}
	if level > DebugLevel {
		debugLog.SetOutput(io.Discard)
	}
	if level > InfoLevel {
		infoLog.SetOutput(io.Discard)
	}
	if level > WarnLevel {
		warnLog.SetOutput(io.Discard)
	}
	if level > ErrorLevel {
		warnLog.SetOutput(io.Discard)
	}
}

func New(text string) (err error) {
	return &ormError{
		text: text,
	}
}
