package worker

import (
	"github.com/ppeble/dreamhost-personal-backup"
)

type backupLogger interface {
	Info(backup.LogEntry)
	Error(backup.LogEntry)
}
