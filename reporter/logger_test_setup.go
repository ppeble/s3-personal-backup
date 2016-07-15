package reporter

type sliceLogger struct {
	messages []string
}

func (l *sliceLogger) Write(b []byte) (n int, err error) {
	msg := string(b[:])
	l.messages = append(l.messages, msg)
	return len(msg), nil
}
