package backup

import (
	"context"
	"errors"

	minio "github.com/minio/minio-go/v7"
)

type RemoteFileProcessor struct {
	bucket   string
	fileData FileData

	list   func(context.Context, string, minio.ListObjectsOptions) <-chan minio.ObjectInfo
	remove func(context.Context, string, string, minio.RemoveObjectOptions) error
	put    func(context.Context, string, string, string, minio.PutObjectOptions) (minio.UploadInfo, error)
}

func NewRemoteFileProcessor(
	b string,
	l func(context.Context, string, minio.ListObjectsOptions) <-chan minio.ObjectInfo,
	r func(context.Context, string, string, minio.RemoveObjectOptions) error,
	p func(context.Context, string, string, string, minio.PutObjectOptions) (minio.UploadInfo, error),
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
	for object := range p.list(context.Background(), p.bucket, minio.ListObjectsOptions{Prefix: "", Recursive: true}) {
		if object.Err != nil {
			return nil, object.Err
		}

		p.fileData[Filename(object.Key)] = newFile(object.Key, object.Size)
	}

	return p.fileData, nil
}

func (p *RemoteFileProcessor) Remove(f string) error {
	return p.remove(context.Background(), p.bucket, f, minio.RemoveObjectOptions{})
}

func (p *RemoteFileProcessor) Put(f string) (err error) {
	if f == "" {
		err = errors.New("'put' error: target file cannot be missing")
		return
	}

	// We ignore the return file info, we don't need it for now
	_, err = p.put(context.Background(), p.bucket, f, f, minio.PutObjectOptions{
		ContentType: "", // A blank will cause the type to be auto-detected by the lib
	})
	return
}
