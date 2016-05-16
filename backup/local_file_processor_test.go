package backup

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestLocalProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(LocalProcessorTestSuite))
}

type LocalProcessorTestSuite struct {
	suite.Suite
	rootDir   string
	processor localFileProcessor
}

func (s *LocalProcessorTestSuite) SetupTest() {
	s.rootDir = s.createTempDir("", "rootDir")
	s.processor = NewLocalFileProcessor(s.rootDir)
}

func (s *LocalProcessorTestSuite) TeardownTest() {
	os.RemoveAll(s.rootDir)
}

func (s *LocalProcessorTestSuite) Test_Process_Error() {
	processor := NewLocalFileProcessor("bad_file_path")
	_, err := processor.Gather()

	s.Require().Error(err)
}

func (s *LocalProcessorTestSuite) Test_Process_SingleDirSingleFile() {
	tempFile := s.createTempFile(s.rootDir, "TEST")

	localFileInfo, err := s.processor.Gather()
	s.Require().NoError(err)

	s.compare(tempFile, localFileInfo)
}

func (s *LocalProcessorTestSuite) Test_Process_SingleDirMultipleFiles() {
	tempFile1 := s.createTempFile(s.rootDir, "TEST1")
	tempFile2 := s.createTempFile(s.rootDir, "TEST2")
	tempFile3 := s.createTempFile(s.rootDir, "TEST3")

	localFileInfo, err := s.processor.Gather()
	s.Require().NoError(err)

	s.compare(tempFile1, localFileInfo)
	s.compare(tempFile2, localFileInfo)
	s.compare(tempFile3, localFileInfo)
}

func (s *LocalProcessorTestSuite) Test_Process_TwoDirsSingleFile() {
	rootDirTempFile := s.createTempFile(s.rootDir, "rootDirTempFile")

	innerTempDir := s.createTempDir(s.rootDir, "innerDir")
	innerTempFile := s.createTempFile(innerTempDir, "innerTestFile")

	localFileInfo, err := s.processor.Gather()
	s.Require().NoError(err)

	s.compare(rootDirTempFile, localFileInfo)
	s.compare(innerTempFile, localFileInfo)
}

func (s *LocalProcessorTestSuite) Test_Process_TwoDirsMultipleFiles() {
	rootDirTempFile1 := s.createTempFile(s.rootDir, "rootDirTempFile1")
	rootDirTempFile2 := s.createTempFile(s.rootDir, "rootDirTempFile2")
	rootDirTempFile3 := s.createTempFile(s.rootDir, "rootDirTempFile3")

	innerTempDir := s.createTempDir(s.rootDir, "innerDir")
	innerTempFile1 := s.createTempFile(innerTempDir, "innerTestFile1")
	innerTempFile2 := s.createTempFile(innerTempDir, "innerTestFile2")
	innerTempFile3 := s.createTempFile(innerTempDir, "innerTestFile3")

	localFileInfo, err := s.processor.Gather()
	s.Require().NoError(err)

	s.compare(rootDirTempFile1, localFileInfo)
	s.compare(rootDirTempFile2, localFileInfo)
	s.compare(rootDirTempFile3, localFileInfo)
	s.compare(innerTempFile1, localFileInfo)
	s.compare(innerTempFile2, localFileInfo)
	s.compare(innerTempFile3, localFileInfo)
}

// Test dir format:
// /rootDir
//   f1
//   /nestedDir1
//     innerF1
//     innerF2
//     innerF3
//   /nestedDir2
//     inner2F1
//     inner2F2
//     /nestedDir3
//       inner3F1
//       /nestedDir4
//         inner4f1
//       /nestedDir5 (this is empty)
func (s *LocalProcessorTestSuite) Test_Process_MultipleDirsMultipleFilesDifferentNestedLevels() {
	rootDirTempFile1 := s.createTempFile(s.rootDir, "rootDirTempFile1")

	nestedTempDir1 := s.createTempDir(s.rootDir, "nestedDir1")
	nestedDir1TempFile1 := s.createTempFile(nestedTempDir1, "nestedDir1TestFile1")
	nestedDir1TempFile2 := s.createTempFile(nestedTempDir1, "nestedDir1TestFile2")
	nestedDir1TempFile3 := s.createTempFile(nestedTempDir1, "nestedDir1TestFile3")

	nestedTempDir2 := s.createTempDir(s.rootDir, "nestedDir2")
	nestedDir2TempFile1 := s.createTempFile(nestedTempDir2, "nestedDir2TestFile1")
	nestedDir2TempFile2 := s.createTempFile(nestedTempDir2, "nestedDir2TestFile2")

	nestedTempDir3 := s.createTempDir(nestedTempDir2, "nestedDir3")
	nestedDir3TempFile1 := s.createTempFile(nestedTempDir3, "nestedDir3TestFile1")

	nestedTempDir4 := s.createTempDir(nestedTempDir3, "nestedDir4")
	nestedDir4TempFile1 := s.createTempFile(nestedTempDir4, "nestedDir4TestFile1")

	s.createTempDir(nestedTempDir3, "nestedDir5")

	localFileInfo, err := s.processor.Gather()
	s.Require().NoError(err)

	s.compare(rootDirTempFile1, localFileInfo)

	s.compare(nestedDir1TempFile1, localFileInfo)
	s.compare(nestedDir1TempFile2, localFileInfo)
	s.compare(nestedDir1TempFile3, localFileInfo)

	s.compare(nestedDir2TempFile1, localFileInfo)
	s.compare(nestedDir2TempFile2, localFileInfo)

	s.compare(nestedDir3TempFile1, localFileInfo)

	s.compare(nestedDir4TempFile1, localFileInfo)
}

func (s *LocalProcessorTestSuite) createTempDir(directory, prefix string) string {
	createdDir, err := ioutil.TempDir(directory, prefix)
	if err != nil {
		s.T().Fatal(err)
	}

	return createdDir
}

func (s *LocalProcessorTestSuite) createTempFile(directory, prefix string) *os.File {
	tmpFile, err := ioutil.TempFile(directory, prefix)
	if err != nil {
		s.T().Fatal(err)
	}

	return tmpFile
}

func (s *LocalProcessorTestSuite) compare(tmpFile *os.File, data map[string]file) {
	fi, err := tmpFile.Stat()
	if err != nil {
		s.T().Fatal(err)
	}

	expected := newFile(tmpFile.Name(), fi.Size())

	actual, found := data[tmpFile.Name()]
	s.True(found)
	s.Equal(expected, actual)
}
