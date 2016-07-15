package worker

import (
	"github.com/ptrimble/dreamhost-personal-backup"
)

type backupLogger interface {
	Info(backup.LogEntry)
	Error(backup.LogEntry)
}
