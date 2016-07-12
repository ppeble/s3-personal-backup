package backup

type backupLogger interface {
	Info(LogEntry)
	Error(LogEntry)
}
