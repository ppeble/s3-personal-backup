package worker

import (
	"github.com/ppeble/s3-personal-backup/pkg/backup"
)

type testLogger struct {
	logInfo  func(backup.LogEntry)
	logError func(backup.LogEntry)
}

func (l testLogger) Info(i backup.LogEntry) {
	l.logInfo(i)
}

func (l testLogger) Error(i backup.LogEntry) {
	l.logError(i)
}
