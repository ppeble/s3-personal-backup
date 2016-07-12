package backup

import (
	"fmt"
)

type LogEntry struct {
	Message, File, Level string
	ActionType           ActionType
}

func (l LogEntry) String() string {
	return fmt.Sprintf("file: '%s' - action: '%s' - message: '%s'", l.File, l.ActionType, l.Message)
}
