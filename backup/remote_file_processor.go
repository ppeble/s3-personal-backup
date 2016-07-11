package backup

import (
	"errors"

	minio "github.com/minio/minio-go"
)

type RemoteFileProcessor struct {
	bucket   string
	fileData FileData

	list   func(string, string, bool, <-chan struct{}) <-chan minio.ObjectInfo
	remove func(string, string) error
	put    func(string, string, string, string) (int64, error)
}

func NewRemoteFileProcessor(
	b string,
	l func(string, string, bool, <-chan struct{}) <-chan minio.ObjectInfo,
	r func(string, string) error,
	p func(string, string, string, string) (int64, error),
) (RemoteFileProcessor, error) {
	if b == "" {
		return RemoteFileProcessor{}, errors.New("'NewRemoteFileProcessor' error: bucket cannot be missing")
	}

	return RemoteFileProcessor{
		bucket:   b,
		list:     l,
		remove:   r,
		put:      p,
		fileData: make(FileData, 0),
	}, nil
}

func (p *RemoteFileProcessor) Gather() (data FileData, err error) {
	// Create a done channel to control 'ListObjects' go routine.
	doneCh := make(chan struct{})

	// Indicate to our routine to exit cleanly upon return.
	defer close(doneCh)

	prefix := ""
	isRecursive := true
	objectCh := p.list(p.bucket, prefix, isRecursive, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		p.fileData[Filename(object.Key)] = newFile(object.Key, object.Size)
	}

	return p.fileData, nil
}

func (p *RemoteFileProcessor) Remove(f string) (err error) {
	err = p.remove(p.bucket, f)
	return
}

func (p *RemoteFileProcessor) Put(f string) (err error) {
	if f == "" {
		err = errors.New("'put' error: target file cannot be missing")
		return
	}

	contentType := "" // A blank will cause the type to be auto-detected by the lib
	_, err = p.put(p.bucket, f, f, contentType)
	return
}
