package controller

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Netflix/go-env"
	"github.com/m-mizutani/catbox/pkg/interfaces"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

type Controller struct {
	Config
	Adaptors

	s3Client  interfaces.S3Client
	sqsClient interfaces.SQSClient
	ecrClient interfaces.ECRClient
}

func New() *Controller {
	ctrl := &Controller{
		Adaptors: defaultAdaptors,
	}

	if err := ctrl.Config.UnmarshalEnvVars(); err != nil {
		panic("Failed unmarshal env vars: " + err.Error())
	}

	return ctrl
}

// Config is structure to unmarshal environment variables for lambda
type Config struct {
	AwsRegion string `env:"AWS_REGION"`

	TableName string `env:"TABLE_NAME"`
	S3Region  string `env:"S3_REGION"`
	S3Bucket  string `env:"S3_BUCKET"`
	S3Prefix  string `env:"S3_PREFIX"`

	ScanQueueURL    string `env:"SCAN_QUEUE_URL"`
	InspectQueueURL string `env:"INSPECT_QUEUE_URL"`
}

// UnmarshalEnvVars retrieves environment variables for lambda
func (x *Config) UnmarshalEnvVars() error {
	if _, err := env.UnmarshalFromEnviron(x); err != nil {
		return golambda.WrapError(err, "env.UnmarshalFromEnviron")
	}

	return nil
}

type Adaptors struct {
	// AWS
	NewS3  interfaces.S3ClientFactory
	NewSQS interfaces.SQSClientFactory
	NewECR interfaces.ECRClientFactory

	// FileSystem
	Exec     interfaces.ExecOutputFunc
	Create   interfaces.FileCreateFunc
	MkdirAll interfaces.MkDirAllFunc
	TempFile interfaces.TempFileFunc
	ReadFile interfaces.ReadFileFunc

	// HTTP
	HTTP interfaces.HTTPClient

	// DB
	NewDBClient interfaces.DBClientFactory
}

// new returns default adaptor set
var defaultAdaptors = Adaptors{
	NewS3:  interfaces.NewS3Client,
	NewSQS: interfaces.NewSQSClient,
	NewECR: interfaces.NewECRClient,

	Exec:     interfaces.DefaultExecOutput,
	Create:   interfaces.DefaultCreateFunc,
	MkdirAll: os.MkdirAll,
	TempFile: interfaces.DefaultTempFileFunc,
	ReadFile: ioutil.ReadFile,

	HTTP: &http.Client{},
}
