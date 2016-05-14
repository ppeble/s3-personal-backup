package backup

import (
	"os"
	"path"
	"path/filepath"
)

type localFileProcessor struct {
	fileData        map[string][]os.FileInfo
	isRootProcessed bool
}

func NewLocalFileProcessor() localFileProcessor {
	return localFileProcessor{
		fileData:        make(map[string][]os.FileInfo, 0),
		isRootProcessed: false,
	}
}

func (p *localFileProcessor) Process(targetDir string) (data map[string][]os.FileInfo, err error) {
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

	var dir string
	if !p.isRootProcessed {
		p.isRootProcessed = true
		dir = filePath
	} else {
		dir = path.Dir(filePath)
	}

	if p.fileData[dir] == nil {
		p.fileData[dir] = make([]os.FileInfo, 0)
	}

	if fi.IsDir() {
		p.fileData[filePath] = make([]os.FileInfo, 0)
	} else {
		p.fileData[dir] = append(p.fileData[dir], fi)
	}

	return
}
