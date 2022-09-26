package worker

import (
	"github.com/ppeble/s3-personal-backup/pkg/backup"
)

type backupLogger interface {
	Info(backup.LogEntry)
	Error(backup.LogEntry)
}
