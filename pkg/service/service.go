package service

import (
	"github.com/m-mizutani/catbox/pkg/interfaces"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

type Service struct {
	config *interfaces.Config

	s3Client  interfaces.S3Client
	sqsClient interfaces.SQSClient
	ecrClient interfaces.ECRClient
}

func New(config *interfaces.Config) *Service {
	return &Service{
		config: config,
	}
}
