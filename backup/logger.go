package backup

import (
	"io"
	"log"
)

const (
	INFO  = "INFO"
	ERROR = "ERROR"
)

type BackupLogger interface {
	Info(LogEntry)
	Error(LogEntry)
}

func NewLogger(out io.Writer, report chan<- LogEntry) BackupLogger {
	return backupLogger{
		infoLog:  log.New(out, INFO+": ", log.Ldate|log.Ltime|log.LUTC),
		errorLog: log.New(out, ERROR+": ", log.Ldate|log.Ltime|log.LUTC),
		report:   report,
	}
}

type backupLogger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	report   chan<- LogEntry
}

type LogEntry struct {
	message, file, level string
}

func (l backupLogger) Info(i LogEntry) {
	l.infoLog.Println("file: '" + i.file + "' - message: '" + i.message + "'")
	l.sendToReporter(i)
}

func (l backupLogger) Error(i LogEntry) {
	l.errorLog.Println("file: '" + i.file + "' - message: '" + i.message + "'")
	l.sendToReporter(i)
}

func (l backupLogger) sendToReporter(i LogEntry) {
	go func() {
		l.report <- i
	}()
}
