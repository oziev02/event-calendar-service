package logger

import (
	"fmt"
	"time"
)

// LogLevel представляет уровень логирования
type LogLevel int

const (
	LevelInfo LogLevel = iota
	LevelError
)

// LogEntry представляет запись лога
type LogEntry struct {
	Level     LogLevel
	Timestamp time.Time
	Message   string
	Fields    map[string]interface{}
}

// Logger интерфейс для асинхронного логирования
type Logger interface {
	Log(level LogLevel, message string, fields map[string]interface{})
	Close() error
}

// AsyncLogger реализует асинхронное логирование через канал
type AsyncLogger struct {
	logChan chan LogEntry
	done    chan struct{}
}

// NewAsyncLogger создает новый асинхронный логгер
func NewAsyncLogger(bufferSize int) *AsyncLogger {
	logger := &AsyncLogger{
		logChan: make(chan LogEntry, bufferSize),
		done:    make(chan struct{}),
	}
	
	go logger.process()
	return logger
}

// Log отправляет запись лога в канал
func (l *AsyncLogger) Log(level LogLevel, message string, fields map[string]interface{}) {
	select {
	case l.logChan <- LogEntry{
		Level:     level,
		Timestamp: time.Now(),
		Message:   message,
		Fields:    fields,
	}:
	default:
		// Канал полон, логировать напрямую, чтобы избежать блокировки
		fmt.Printf("[FALLBACK] %s: %s\n", levelString(level), message)
	}
}

// Close останавливает логгер
func (l *AsyncLogger) Close() error {
	close(l.logChan)
	<-l.done
	return nil
}

// process обрабатывает записи логов из канала
func (l *AsyncLogger) process() {
	defer close(l.done)
	
	for entry := range l.logChan {
		timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
		fieldsStr := ""
		if len(entry.Fields) > 0 {
			fieldsStr = " "
			for k, v := range entry.Fields {
				fieldsStr += fmt.Sprintf("%s=%v ", k, v)
			}
		}
		fmt.Printf("[%s] [%s] %s%s\n", timestamp, levelString(entry.Level), entry.Message, fieldsStr)
	}
}

func levelString(level LogLevel) string {
	switch level {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

