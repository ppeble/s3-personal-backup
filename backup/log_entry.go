package backup

import (
	"fmt"
)

type LogEntry struct {
	Message, File, Level string
}

func (l LogEntry) String() string {
	return fmt.Sprintf("file: '%s' - message: '%s'", l.File, l.Message)
}
