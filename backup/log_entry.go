package backup

import (
	"fmt"
)

type LogEntry struct {
	message, file, level string
}

//TODO This should probably have a matching test for the sake of completeness
func (l LogEntry) String() string {
	return fmt.Sprintf("file: '%s' - message: '%s'", l.file, l.message)
}
