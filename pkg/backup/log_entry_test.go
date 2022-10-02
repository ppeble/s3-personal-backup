package backup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LogEntry_String_Push(t *testing.T) {
	entry := LogEntry{
		Message:    "Message",
		File:       "File",
		Level:      "Level",
		ActionType: PUSH,
	}

	assert.Equal(t, "file: 'File' - action: 'push' - message: 'Message'", entry.String())
}

func Test_LogEntry_String_Remove(t *testing.T) {
	entry := LogEntry{
		Message:    "Message",
		File:       "File",
		Level:      "Level",
		ActionType: REMOVE,
	}

	assert.Equal(t, "file: 'File' - action: 'remove' - message: 'Message'", entry.String())
}
