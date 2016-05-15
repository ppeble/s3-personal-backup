package backup

import (
	"errors"
	"testing"

	minio "github.com/minio/minio-go"
	"github.com/stretchr/testify/suite"
)

func TestRemoteProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(RemoteProcessorTestSuite))
}

type RemoteProcessorTestSuite struct {
	suite.Suite
	bucketName string
}

func (s *RemoteProcessorTestSuite) SetupTest() {
	s.bucketName = "testBucket"
}

func (s *RemoteProcessorTestSuite) Test_Process_CallsListRemoteObjects() {
	called := false

	listObjectsFunc := func(bucket, prefix string, isRecursive bool, doneCh chan struct{}) <-chan minio.ObjectInfo {
		objectCh := make(chan minio.ObjectInfo, 1)
		defer close(objectCh)

		s.Equal(s.bucketName, bucket)
		s.Equal("", prefix)
		s.Equal(true, isRecursive)

		called = true

		return objectCh
	}

	processor := NewRemoteFileProcessor(listObjectsFunc)
	_, err := processor.Gather(s.bucketName)

	s.Require().NoError(err)
	s.True(called)
}

func (s *RemoteProcessorTestSuite) Test_Process_ErrorBlankBucketName() {
	expectedErr := errors.New("asplode")

	listObjectsFuncWithError := func(bucket, prefix string, isRecursive bool, doneCh chan struct{}) <-chan minio.ObjectInfo {
		//FIXME I don't understand how this works in the minio example? The
		// error case does basically what I am doing but I thought that senders
		// blocked until a receiver was present. How does this work unbuffered?
		objectCh := make(chan minio.ObjectInfo, 1)
		defer close(objectCh)

		objectCh <- minio.ObjectInfo{
			Err: expectedErr,
		}

		return objectCh
	}

	processor := NewRemoteFileProcessor(listObjectsFuncWithError)
	_, err := processor.Gather("")

	s.Require().Error(err)
	s.Equal(expectedErr, err)
}

func (s *RemoteProcessorTestSuite) Test_Process_SingleFile() {
	listObjectsFunc := func(bucket, prefix string, isRecursive bool, doneCh chan struct{}) <-chan minio.ObjectInfo {
		objectCh := make(chan minio.ObjectInfo, 1)
		defer close(objectCh)

		objectCh <- minio.ObjectInfo{
			Key:  "test",
			Size: 100,
		}

		return objectCh
	}

	processor := NewRemoteFileProcessor(listObjectsFunc)
	data, err := processor.Gather(s.bucketName)

	s.Require().NoError(err)
	s.Equal(newFile("test", 100), data["test"])
}

func (s *RemoteProcessorTestSuite) Test_Process_MultipleFiles() {
	listObjectsFunc := func(bucket, prefix string, isRecursive bool, doneCh chan struct{}) <-chan minio.ObjectInfo {
		objectCh := make(chan minio.ObjectInfo, 3)
		defer close(objectCh)

		objectCh <- minio.ObjectInfo{
			Key:  "test1",
			Size: 100,
		}

		objectCh <- minio.ObjectInfo{
			Key:  "test2",
			Size: 500,
		}

		objectCh <- minio.ObjectInfo{
			Key:  "test3",
			Size: 1000,
		}

		return objectCh
	}

	processor := NewRemoteFileProcessor(listObjectsFunc)
	data, err := processor.Gather(s.bucketName)

	s.Require().NoError(err)
	s.Equal(newFile("test1", 100), data["test1"])
	s.Equal(newFile("test2", 500), data["test2"])
	s.Equal(newFile("test3", 1000), data["test3"])
}
