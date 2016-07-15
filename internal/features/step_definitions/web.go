package step_definitions

import (
	"fmt"
	"time"

	"github.com/gorilla/mux"
	. "github.com/lsegal/gucumber"
	"github.com/minio/minio-go"
	"github.com/stretchr/testify/assert"
)

var SetupWebSteps = setupWebSteps()
var s3Client *minio.Client

func setupWebSteps() bool {
	Before("", func() {
		s3Client, err := minio.NewV2(
			viper.GetString("s3Host"),
			viper.GetString("s3AccessKey"),
			viper.GetString("s3SecretKey"),
			false,
		)
		if err != nil {
			panic(err)
		}
	})
}
