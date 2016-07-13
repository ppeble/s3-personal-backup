package reporter

import (
	"log"

	"github.com/ptrimble/dreamhost-personal-backup/backup"
)

type dryRunReporter struct {
	in     <-chan backup.LogEntry
	logger *log.Logger

	entries []backup.LogEntry

	pushCount, removeCount int
}

func NewDryRunReporter(
	in <-chan backup.LogEntry,
	l *log.Logger,
) dryRunReporter {
	return dryRunReporter{
		in:          in,
		logger:      l,
		entries:     make([]backup.LogEntry, 0),
		pushCount:   0,
		removeCount: 0,
	}
}

func (r *dryRunReporter) Run() {
	for {
		entry := <-r.in
		r.entries = append(r.entries, entry)

		if entry.ActionType == backup.PUSH {
			r.pushCount++
		} else if entry.ActionType == backup.REMOVE {
			r.removeCount++
		}
	}
}

func (r *dryRunReporter) Print() {
	r.logger.Println("Dry Run Report")
	r.logger.Println("-------------------------------")
	r.logger.Printf("Total files processed: %d\n", len(r.entries))
	r.logger.Printf("Files that would be added to remote: %d\n", r.pushCount)
	r.logger.Printf("Files that would be removed from remote: %d\n", r.removeCount)
	r.logger.Println("")
	r.logger.Println("File Details")
	r.logger.Println("-------------------------------")

	for _, entry := range r.entries {
		r.logger.Println(entry.String())
	}

	r.logger.Println("")
}
