package backup

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

type ConfigTestSuite struct {
	suite.Suite
	flags Flags
}

func (s *ConfigTestSuite) SetupTest() {
	s.flags = Flags{
		TargetDir:    "targetDir",
		S3Host:       "s3Host",
		S3AccessKey:  "s3AccessKey",
		S3SecretKey:  "s3SecretKey",
		S3BucketName: "s3BucketName",
	}
}

func (s *ConfigTestSuite) Test_TakesAllFlags() {
	c, err := CompileConfig(s.flags)

	s.NoError(err)
	s.Equal(s.flags.TargetDir, c.TargetDir)
	s.Equal(s.flags.S3Host, c.S3Host)
	s.Equal(s.flags.S3AccessKey, c.S3AccessKey)
	s.Equal(s.flags.S3SecretKey, c.S3SecretKey)
	s.Equal(s.flags.S3BucketName, c.S3BucketName)
}

func (s *ConfigTestSuite) Test_TargetDir_TakesEnv() {
	os.Setenv("PERSONAL_BACKUP_TARGET_DIR", "OS_TARGET_DIR")

	s.flags.TargetDir = ""
	c, err := CompileConfig(s.flags)

	s.NoError(err)
	s.Equal("OS_TARGET_DIR", c.TargetDir)
}

func (s *ConfigTestSuite) Test_TargetDir_Missing() {
	s.flags.TargetDir = ""

	_, err := CompileConfig(s.flags)

	s.Error(err)
}

func (s *ConfigTestSuite) Test_S3Host_TakesEnv() {
	os.Setenv("PERSONAL_BACKUP_S3_HOST", "OS_S3_HOST")

	s.flags.S3Host = ""
	c, err := CompileConfig(s.flags)

	s.NoError(err)
	s.Equal("OS_S3_HOST", c.S3Host)
}

func (s *ConfigTestSuite) Test_S3Host_Missing() {
	s.flags.S3Host = ""

	_, err := CompileConfig(s.flags)

	s.Error(err)
}

func (s *ConfigTestSuite) Test_S3AccessKey_TakesEnv() {
	os.Setenv("PERSONAL_BACKUP_S3_ACCESS_KEY", "OS_S3_ACCESS_KEY")

	s.flags.S3AccessKey = ""
	c, err := CompileConfig(s.flags)

	s.NoError(err)
	s.Equal("OS_S3_ACCESS_KEY", c.S3AccessKey)
}

func (s *ConfigTestSuite) Test_S3AccessKey_Missing() {
	s.flags.S3AccessKey = ""

	_, err := CompileConfig(s.flags)

	s.Error(err)
}

func (s *ConfigTestSuite) Test_S3SecretKey_TakesEnv() {
	os.Setenv("PERSONAL_BACKUP_S3_SECRET_KEY", "OS_S3_SECRET_KEY")

	s.flags.S3SecretKey = ""
	c, err := CompileConfig(s.flags)

	s.NoError(err)
	s.Equal("OS_S3_SECRET_KEY", c.S3SecretKey)
}

func (s *ConfigTestSuite) Test_S3SecretKey_Missing() {
	s.flags.S3SecretKey = ""

	_, err := CompileConfig(s.flags)

	s.Error(err)
}

func (s *ConfigTestSuite) Test_S3BucketName_TakesEnv() {
	os.Setenv("PERSONAL_BACKUP_S3_BUCKET_NAME", "OS_S3_BUCKET_NAME")

	s.flags.S3BucketName = ""
	c, err := CompileConfig(s.flags)

	s.NoError(err)
	s.Equal("OS_S3_BUCKET_NAME", c.S3BucketName)
}

func (s *ConfigTestSuite) Test_S3BucketName_Missing() {
	s.flags.S3BucketName = ""

	_, err := CompileConfig(s.flags)

	s.Error(err)
}
