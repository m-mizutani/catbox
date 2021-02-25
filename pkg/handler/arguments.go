package handler

import (
	env "github.com/Netflix/go-env"
	"github.com/m-mizutani/golambda"

	"github.com/m-mizutani/catbox/pkg/adaptor"
	"github.com/m-mizutani/catbox/pkg/db"
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

// Arguments is settings of lambda without event
type Arguments struct {
	EnvVars

	Adaptors *adaptor.Adaptors
	DBClient db.DBClient
}

// NewArguments sets up Arguments to unmarshal EnvVars
func NewArguments() (*Arguments, error) {
	args := &Arguments{
		Adaptors: adaptor.New(),
	}

	if err := args.EnvVars.Unmarshal(); err != nil {
		return nil, err
	}

	return args, nil
}
