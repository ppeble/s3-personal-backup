package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LogEntry_String(t *testing.T) {
	entry := LogEntry{
		Message: "Message",
		File:    "File",
		Level:   "Level",
	}

	assert.Equal(t, "file: 'File' - message: 'Message'", entry.String())
}
