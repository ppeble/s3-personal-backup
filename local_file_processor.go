package backup

import (
	"os"
	"path/filepath"
)

type localFileProcessor struct {
	fileData map[string]file
}

func NewLocalFileProcessor() localFileProcessor {
	return localFileProcessor{
		fileData: make(map[string]file, 0),
	}
}

func (p *localFileProcessor) Gather(targetDir string) (data map[string]file, err error) {
	err = filepath.Walk(targetDir, p.processFile)
	if err != nil {
		return
	}

	data = p.fileData

	return
}

func (p *localFileProcessor) processFile(filePath string, fi os.FileInfo, err error) (e error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return err
	}

	if !fi.IsDir() {
		p.fileData[filePath] = newFile(filePath, fi.Size())
	}

	return
}
