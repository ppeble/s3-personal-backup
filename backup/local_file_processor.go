package backup

import (
	"os"
	"path/filepath"
)

type LocalFileProcessor struct {
	targetDir string
	fileData  FileData
}

//FIXME This should return an error if the target is blank/missing
func NewLocalFileProcessor(t string) LocalFileProcessor {
	return LocalFileProcessor{
		targetDir: t,
		fileData:  make(FileData),
	}
}

func (p *LocalFileProcessor) Gather() (data FileData, err error) {
	err = filepath.Walk(p.targetDir, p.processFile)
	if err != nil {
		return
	}

	data = p.fileData

	return
}

func (p *LocalFileProcessor) processFile(filePath string, fi os.FileInfo, err error) (e error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return err
	}

	if !fi.IsDir() {
		p.fileData[Filename(filePath)] = newFile(filePath, fi.Size())
	}

	return
}
