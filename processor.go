package backup

import (
	"os"
)

// Make a constructor that accepts methods to call
// This way we can test this glue package
// Then in main we pass in real things

type processor struct {
	localFileProcessor func(string) (map[string][]os.FileInfo, error)
	// etc
}

func NewProcessor(local func(string) (map[string][]os.FileInfo, error)) processor {
	return processor{
		localFileProcessor: local,
	}
}

func (p processor) Process(targetDir string) error {
	_, err := p.localFileProcessor(targetDir)
	return err
}
