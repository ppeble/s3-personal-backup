package backup

import (
	"os"
	"path/filepath"
)

type localFileProcessor struct {
	targetDir string
	fileData  map[string]file
}

//FIXME This should return an error if the target is blank/missing
func NewLocalFileProcessor(t string) localFileProcessor {
	return localFileProcessor{
		targetDir: t,
		fileData:  make(map[string]file, 0),
	}
}

func (p *localFileProcessor) Gather() (data map[string]file, err error) {
	err = filepath.Walk(p.targetDir, p.processFile)
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
