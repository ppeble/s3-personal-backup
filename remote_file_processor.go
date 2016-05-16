package backup

import (
	"errors"

	minio "github.com/minio/minio-go"
)

type remoteFileProcessor struct {
	list     func(string, string, bool, <-chan struct{}) <-chan minio.ObjectInfo
	remove   func(string, string) error
	put      func(string, string, string, string) (int64, error)
	fileData map[string]file
}

func NewRemoteFileProcessor(l func(string, string, bool, <-chan struct{}) <-chan minio.ObjectInfo, r func(string, string) error, p func(string, string, string, string) (int64, error)) remoteFileProcessor {
	return remoteFileProcessor{
		list:     l,
		remove:   r,
		put:      p,
		fileData: make(map[string]file, 0),
	}
}

func (p *remoteFileProcessor) Gather(bucket string) (data map[string]file, err error) {
	// Create a done channel to control 'ListObjects' go routine.
	doneCh := make(chan struct{})

	// Indicate to our routine to exit cleanly upon return.
	defer close(doneCh)

	prefix := ""
	isRecursive := true
	objectCh := p.list(bucket, prefix, isRecursive, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		p.fileData[object.Key] = newFile(object.Key, object.Size)
	}

	return p.fileData, nil
}

func (p *remoteFileProcessor) Remove(b, fn string) (err error) {
	err = p.remove(b, fn)
	return
}

func (p *remoteFileProcessor) Put(b, f string) (err error) {
	if f == "" {
		err = errors.New("'put' error: target file cannot be missing")
		return
	}

	contentType := "" // A blank will cause the type to be auto-detected by the lib
	_, err = p.put(b, f, f, contentType)
	return
}
