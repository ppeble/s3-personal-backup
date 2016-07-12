package worker

import (
	"github.com/ptrimble/dreamhost-personal-backup/backup"
)

type backupLogger interface {
	Info(backup.LogEntry)
	Error(backup.LogEntry)
}
