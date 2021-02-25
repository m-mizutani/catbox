package adaptor

import (
	"io/ioutil"
	"net/http"
	"os"
)

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
}

// New returns default adaptor set
func New() *Adaptors {
	return &Adaptors{
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
}
