package worker

import (
	"github.com/ptrimble/dreamhost-personal-backup/backup/logger"
)

type testLogger struct {
	logInfo  func(logger.LogEntry)
	logError func(logger.LogEntry)
}

func (l testLogger) Info(i logger.LogEntry) {
	l.logInfo(i)
}

func (l testLogger) Error(i logger.LogEntry) {
	l.logError(i)
}
