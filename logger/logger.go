package logger

import (
	"io"
	"log"
	"sync"

	"github.com/ptrimble/dreamhost-personal-backup"
)

const (
	INFO  = "INFO"
	ERROR = "ERROR"
)

func NewLogger(out io.Writer, report chan<- backup.LogEntry, wg *sync.WaitGroup) backupLogger {
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
	report   chan<- backup.LogEntry
	wg       *sync.WaitGroup
}

//FIXME Can't we just print the log entry? Why not? Why do it again here?
func (l backupLogger) Info(i backup.LogEntry) {
	l.infoLog.Println("file: '" + i.File + "' - message: '" + i.Message + "'")
	l.sendToReporter(i)
}

//FIXME Can't we just print the log entry? Why not? Why do it again here?
func (l backupLogger) Error(i backup.LogEntry) {
	l.errorLog.Println("file: '" + i.File + "' - message: '" + i.Message + "'")
	l.sendToReporter(i)
}

func (l backupLogger) sendToReporter(i backup.LogEntry) {
	l.wg.Add(1)
	go func() {
		l.report <- i
		l.wg.Done()
	}()
}
