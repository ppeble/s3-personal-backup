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
	bucket     string
	listFunc   func(string, string, bool, <-chan struct{}) <-chan minio.ObjectInfo
	removeFunc func(string, string) error
	putFunc    func(string, string, string, string) (int64, error)
}

func (s *RemoteProcessorTestSuite) SetupTest() {
	s.bucket = "testBucket"

	s.listFunc = func(bucket, prefix string, isRecursive bool, doneCh <-chan struct{}) <-chan minio.ObjectInfo {
		objectCh := make(chan minio.ObjectInfo)
		defer close(objectCh)
		return objectCh
	}

	s.removeFunc = func(bucket, file string) error { return nil }
	s.putFunc = func(bucket, file, filepath, contentType string) (int64, error) { return 0, nil }
}

func (s *RemoteProcessorTestSuite) Test_Gather_CallsListRemoteObjects() {
	called := false

	listFunc := func(bucket, prefix string, isRecursive bool, doneCh <-chan struct{}) <-chan minio.ObjectInfo {
		objectCh := make(chan minio.ObjectInfo, 1)
		defer close(objectCh)

		s.Equal(s.bucket, bucket)
		s.Equal("", prefix)
		s.Equal(true, isRecursive)

		called = true

		return objectCh
	}

	processor, _ := NewRemoteFileProcessor(s.bucket, listFunc, s.removeFunc, s.putFunc)
	_, err := processor.Gather()

	s.Require().NoError(err)
	s.True(called)
}

func (s *RemoteProcessorTestSuite) Test_Gather_ReturnsErrorOnlyForFailedObjects() {
	called := false
	expectedErr := errors.New("asplode")

	listFunc := func(bucket, prefix string, isRecursive bool, doneCh <-chan struct{}) <-chan minio.ObjectInfo {
		objectCh := make(chan minio.ObjectInfo, 1)
		defer close(objectCh)

		objectCh <- minio.ObjectInfo{
			Err: expectedErr,
		}

		called = true
		return objectCh
	}

	processor, _ := NewRemoteFileProcessor(s.bucket, listFunc, s.removeFunc, s.putFunc)
	_, err := processor.Gather()

	s.Require().Error(err)
	s.True(called)
	s.Equal(expectedErr, err)
}

func (s *RemoteProcessorTestSuite) Test_New_ErrorBlankBucketName() {
	_, err := NewRemoteFileProcessor("", s.listFunc, s.removeFunc, s.putFunc)
	s.Error(err)
	s.Equal(errors.New("'NewRemoteFileProcessor' error: bucket cannot be missing"), err)
}

func (s *RemoteProcessorTestSuite) Test_Gather_SingleFile() {
	listFunc := func(bucket, prefix string, isRecursive bool, doneCh <-chan struct{}) <-chan minio.ObjectInfo {
		objectCh := make(chan minio.ObjectInfo, 1)
		defer close(objectCh)

		objectCh <- minio.ObjectInfo{
			Key:  "test",
			Size: 100,
		}

		return objectCh
	}

	processor, _ := NewRemoteFileProcessor(s.bucket, listFunc, s.removeFunc, s.putFunc)
	data, err := processor.Gather()

	s.Require().NoError(err)
	s.Equal(newFile("test", 100), data["test"])
}

func (s *RemoteProcessorTestSuite) Test_Gather_MultipleFiles() {
	listFunc := func(bucket, prefix string, isRecursive bool, doneCh <-chan struct{}) <-chan minio.ObjectInfo {
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

	processor, _ := NewRemoteFileProcessor(s.bucket, listFunc, s.removeFunc, s.putFunc)
	data, err := processor.Gather()

	s.Require().NoError(err)
	s.Equal(newFile("test1", 100), data["test1"])
	s.Equal(newFile("test2", 500), data["test2"])
	s.Equal(newFile("test3", 1000), data["test3"])
}

func (s *RemoteProcessorTestSuite) Test_Remove_Happy() {
	called := false
	removeFunc := func(bucket, file string) error {
		s.Equal(s.bucket, bucket)
		s.Equal("test", file)
		called = true
		return nil
	}

	processor, _ := NewRemoteFileProcessor(s.bucket, s.listFunc, removeFunc, s.putFunc)
	err := processor.Remove("test")

	s.Require().NoError(err)
	s.True(called)
}

func (s *RemoteProcessorTestSuite) Test_Remove_Error() {
	called := false
	expectedErr := errors.New("asplode")
	removeFunc := func(bucket, file string) error {
		called = true
		return expectedErr
	}

	processor, _ := NewRemoteFileProcessor(s.bucket, s.listFunc, removeFunc, s.putFunc)
	err := processor.Remove("test")

	s.Error(err)
	s.True(called)
	s.Equal(expectedErr, err)
}

func (s *RemoteProcessorTestSuite) Test_Put_Happy() {
	called := false
	expectedFile := "/tmp/test"

	putFunc := func(bucket, fileName, filePath, contentType string) (int64, error) {
		s.Equal(s.bucket, bucket)
		s.Equal(expectedFile, fileName)
		s.Equal(expectedFile, filePath)
		s.Equal("", contentType)

		called = true
		return 0, nil
	}

	processor, _ := NewRemoteFileProcessor(s.bucket, s.listFunc, s.removeFunc, putFunc)

	err := processor.Put(expectedFile)

	s.Require().NoError(err)
	s.True(called)
}

func (s *RemoteProcessorTestSuite) Test_Put_ReturnsErrorOnFailure() {
	called := false
	expectedFile := "/tmp/test"
	expectedErr := errors.New("asplode")

	putFunc := func(bucket, fileName, filePath, contentType string) (int64, error) {
		called = true
		return 0, expectedErr
	}

	processor, _ := NewRemoteFileProcessor(s.bucket, s.listFunc, s.removeFunc, putFunc)

	err := processor.Put(expectedFile)

	s.Error(err)
	s.True(called)
	s.Equal(expectedErr, err)
}

func (s *RemoteProcessorTestSuite) Test_Put_ReturnsErrorIfFileIsMissing() {
	called := false
	expectedFile := ""
	expectedErr := errors.New("'put' error: target file cannot be missing")

	putFunc := func(bucket, fileName, filePath, contentType string) (int64, error) {
		called = true
		return 0, nil
	}

	processor, _ := NewRemoteFileProcessor(s.bucket, s.listFunc, s.removeFunc, putFunc)

	err := processor.Put(expectedFile)

	s.Error(err)
	s.False(called)
	s.Equal(expectedErr, err)
}
