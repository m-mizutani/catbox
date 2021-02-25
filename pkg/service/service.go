package service

import (
	"github.com/m-mizutani/catbox/pkg/adaptor"
	"github.com/m-mizutani/catbox/pkg/handler"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

type Service struct {
	args *handler.Arguments

	s3Client  adaptor.S3Client
	sqsClient adaptor.SQSClient
	ecrClient adaptor.ECRClient
}

func New(args *handler.Arguments) *Service {
	return &Service{
		args: args,
	}
}
