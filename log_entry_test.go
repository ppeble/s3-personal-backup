package backup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LogEntry_String(t *testing.T) {
	entry := LogEntry{
		Message:    "Message",
		File:       "File",
		Level:      "Level",
		ActionType: PUSH,
	}

	assert.Equal(t, "file: 'File' - action: 'push' - message: 'Message'", entry.String())
}
