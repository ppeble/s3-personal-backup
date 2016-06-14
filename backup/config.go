package backup

import (
	"errors"
	"os"
)

type Flags struct {
	TargetDir    string
	S3Host       string
	S3AccessKey  string
	S3SecretKey  string
	S3BucketName string
}

type CompiledConfig struct {
	TargetDir string

	S3Host, S3AccessKey, S3SecretKey, S3BucketName string
}

func CompileConfig(flags Flags) (CompiledConfig, error) {
	c := CompiledConfig{}

	targetDirViaEnv := os.Getenv("PERSONAL_BACKUP_TARGET_DIR")
	s3HostViaEnv := os.Getenv("PERSONAL_BACKUP_S3_HOST")
	s3AccessKeyViaEnv := os.Getenv("PERSONAL_BACKUP_S3_ACCESS_KEY")
	s3SecretKeyViaEnv := os.Getenv("PERSONAL_BACKUP_S3_SECRET_KEY")
	s3BucketNameViaEnv := os.Getenv("PERSONAL_BACKUP_S3_BUCKET_NAME")

	if flags.TargetDir != "" {
		c.TargetDir = flags.TargetDir
	} else if targetDirViaEnv != "" {
		c.TargetDir = targetDirViaEnv
	} else {
		return c, errors.New("target dir must be specified via either command line (-targetDir) or env var (PERSONAL_BACKUP_TARGET_DIR)")
	}

	if flags.S3Host != "" {
		c.S3Host = flags.S3Host
	} else if s3HostViaEnv != "" {
		c.S3Host = s3HostViaEnv
	} else {
		return c, errors.New("s3 host must be specified via either command line (-s3Host) or env var (PERSONAL_BACKUP_S3_HOST)")
	}

	if flags.S3AccessKey != "" {
		c.S3AccessKey = flags.S3AccessKey
	} else if s3AccessKeyViaEnv != "" {
		c.S3AccessKey = s3AccessKeyViaEnv
	} else {
		return c, errors.New("s3 access key must be specified via either command line (-s3AccessKey) or env var (PERSONAL_BACKUP_S3_ACCESS_KEY)")
	}

	if flags.S3SecretKey != "" {
		c.S3SecretKey = flags.S3SecretKey
	} else if s3SecretKeyViaEnv != "" {
		c.S3SecretKey = s3SecretKeyViaEnv
	} else {
		return c, errors.New("s3 secret key must be specified via either command line (-s3SecretKey) or env var (PERSONAL_BACKUP_S3_SECRET_KEY)")
	}

	if flags.S3BucketName != "" {
		c.S3BucketName = flags.S3BucketName
	} else if s3BucketNameViaEnv != "" {
		c.S3BucketName = s3BucketNameViaEnv
	} else {
		return c, errors.New("s3 bucket name must be specified via either command line (-s3BucketName) or env var (PERSONAL_BACKUP_S3_BUCKET_NAME)")
	}

	return c, nil
}
