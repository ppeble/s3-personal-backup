package backup

import (
	"os"
	"path/filepath"
)

type localFileProcessor struct {
	targetDir string
	fileData  map[Filename]File
}

//FIXME This should return an error if the target is blank/missing
func NewLocalFileProcessor(t string) localFileProcessor {
	return localFileProcessor{
		targetDir: t,
		fileData:  make(map[Filename]File, 0),
	}
}

func (p *localFileProcessor) Gather() (data map[Filename]File, err error) {
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
		p.fileData[Filename(filePath)] = newFile(filePath, fi.Size())
	}

	return
}
