package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type DirectoryProcessorTestSuite struct {
	suite.Suite
	directory     string
	tempFile      os.FileInfo
	directoryChan chan []os.FileInfo
}

func (s *DirectoryProcessorTestSuite) SetupTest() {
	createdDir, err := ioutil.TempDir("", "directoryProcessorTemp")
	if err != nil {
		s.T().Fatal(err)
	}

	s.directory = createdDir
	tmpFile, err := ioutil.TempFile(s.directory, "TEST")
	if err != nil {
		s.T().Fatal(err)
	}

	s.tempFile, err = tmpFile.Stat()
	if err != nil {
		s.T().Fatal(err)
	}

	s.directoryChan = make(chan []os.FileInfo)
}

func (s *DirectoryProcessorTestSuite) TeardownTest() {
	os.RemoveAll(s.directory)
}

func (s *DirectoryProcessorTestSuite) Test_ReadsDirAndSendsSliceOfFileInfos() {
	go processDirectories(s.directoryChan, s.directory)

	result := <-s.directoryChan

	s.Equal(s.tempFile, result[0])
}

func TestDirectoryProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(DirectoryProcessorTestSuite))
}
