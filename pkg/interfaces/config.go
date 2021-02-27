package interfaces

import (
	"io/ioutil"
	"net/http"
	"os"

	env "github.com/Netflix/go-env"
	"github.com/m-mizutani/golambda"
)

// EnvVars is structure to unmarshal environment variables for lambda
type EnvVars struct {
	AwsRegion string `env:"AWS_REGION"`

	TableName string `env:"TABLE_NAME"`
	S3Region  string `env:"S3_REGION"`
	S3Bucket  string `env:"S3_BUCKET"`
	S3Prefix  string `env:"S3_PREFIX"`

	ScanQueueURL    string `env:"SCAN_QUEUE_URL"`
	InspectQueueURL string `env:"INSPECT_QUEUE_URL"`
}

// Unmarshal retrieves environment variables for lambda
func (x *EnvVars) Unmarshal() error {
	if _, err := env.UnmarshalFromEnviron(x); err != nil {
		return golambda.WrapError(err, "env.UnmarshalFromEnviron")
	}

	return nil
}

type Adaptors struct {
	// AWS
	NewS3  S3ClientFactory
	NewSQS SQSClientFactory
	NewECR ECRClientFactory

	// FileSystem
	Exec     ExecOutputFunc
	Create   FileCreateFunc
	MkdirAll MkDirAllFunc
	TempFile TempFileFunc
	ReadFile ReadFileFunc

	// HTTP
	HTTP HTTPClient

	// DB
	NewDBClient DBClientFactory
}

// new returns default adaptor set
var defaultAdaptors = Adaptors{
	NewS3:  NewS3Client,
	NewSQS: NewSQSClient,
	NewECR: NewECRClient,

	Exec:     DefaultExecOutput,
	Create:   DefaultCreateFunc,
	MkdirAll: os.MkdirAll,
	TempFile: DefaultTempFileFunc,
	ReadFile: ioutil.ReadFile,

	HTTP: &http.Client{},
}

// Config is settings of lambda without event
type Config struct {
	EnvVars
	Adaptors
}

// NewConfig sets up Config to unmarshal EnvVars
func NewConfig() (*Config, error) {
	args := &Config{
		Adaptors: defaultAdaptors,
	}

	if err := args.EnvVars.Unmarshal(); err != nil {
		return nil, err
	}

	return args, nil
}
