package backup

import (
	"io"
	"log"
	"sync"
)

const (
	INFO  = "INFO"
	ERROR = "ERROR"
)

type BackupLogger interface {
	Info(LogEntry)
	Error(LogEntry)
}

func NewLogger(out io.Writer, report chan<- LogEntry, wg *sync.WaitGroup) BackupLogger {
	return backupLogger{
		infoLog:  log.New(out, INFO+": ", log.Ldate|log.Ltime|log.LUTC),
		errorLog: log.New(out, ERROR+": ", log.Ldate|log.Ltime|log.LUTC),
		report:   report,
		wg:       wg,
	}
}

type backupLogger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	report   chan<- LogEntry
	wg       *sync.WaitGroup
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
	l.wg.Add(1)
	go func() {
		l.report <- i
		l.wg.Done()
	}()
}
