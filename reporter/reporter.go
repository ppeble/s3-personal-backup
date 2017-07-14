package reporter

import (
	"log"
	"time"

	"github.com/ppeble/dreamhost-personal-backup"
)

type reporter struct {
	in     <-chan backup.LogEntry
	logger *log.Logger

	entries []backup.LogEntry
	start   time.Time

	pushCount, removeCount int
}

func NewReporter(
	in <-chan backup.LogEntry,
	l *log.Logger,
) reporter {
	return reporter{
		in:          in,
		logger:      l,
		entries:     make([]backup.LogEntry, 0),
		start:       time.Now(),
		pushCount:   0,
		removeCount: 0,
	}
}

func (r *reporter) Run() {
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

//TODO Add some kind of timestamp in here, this is what we will probably want to be
// printed to a separate file, it'll be nice to have some indication
func (r *reporter) Print() {
	runDuration := time.Since(r.start)
	timePerFile := runDuration.Seconds() / float64(len(r.entries))

	r.logger.Println("Backup Report")
	r.logger.Println("-------------------------------")
	r.logger.Printf("Total run time (in minutes): %d\n", int(runDuration.Minutes()))
	r.logger.Printf("Total files processed: %d\n", len(r.entries))
	r.logger.Printf("Time per file (in seconds): %.4f\n", timePerFile)
	r.logger.Printf("Files added to remote: %d\n", r.pushCount)
	r.logger.Printf("Files removed from remote: %d\n", r.removeCount)
	r.logger.Println("")
	r.logger.Println("File Details")
	r.logger.Println("-------------------------------")

	for _, entry := range r.entries {
		r.logger.Println(entry.String())
	}

	r.logger.Println("")
}
