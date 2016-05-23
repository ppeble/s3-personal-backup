package backup

import (
	"log"
)

func NewReporter(in <-chan logEntry, done <-chan struct{}, l *log.Logger) reporter {
	return reporter{
		in:     in,
		done:   done,
		logger: l,
	}
}

type logEntry struct {
	message, file string
}

type reporter struct {
	in     <-chan logEntry
	done   <-chan struct{}
	logger *log.Logger
}

func (r *reporter) Run() {
	for {
		select {
		case entry := <-r.in:
			r.logger.Println("file: '" + entry.file + "' - message: '" + entry.message + "'")
		case <-r.done:
			r.logger.Println("Received done signal, stopping reporting process")
			return
		}
	}
}
