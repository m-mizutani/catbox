package controller

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

// Hint: Not thread safe
func (x *Controller) setupSQSClient() error {
	if x.sqsClient != nil {
		return nil
	}

	// catbox sends to only SQS queue in same region with Lambda.
	sqsClient, err := x.adaptors.NewSQS(x.AwsRegion)
	if err != nil {
		return golambda.WrapError(err, "Failed to create SQS client").With("region", x.AwsRegion)
	}

	x.sqsClient = sqsClient
	return nil
}

func (x *Controller) sendSQSMessage(url string, data interface{}) error {
	// Run setup only leveraged to avoid high frequency AssumeRole call
	if err := x.setupSQSClient(); err != nil {
		return err
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return golambda.WrapError(err, "json.Marshal SQS").With("data", data)
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    &url,
		MessageBody: aws.String(string(raw)),
	}
	if _, err := x.sqsClient.SendMessage(input); err != nil {
		return golambda.WrapError(err, "sqs.SendMessage").With("input", input)
	}

	return nil
}

func (x *Controller) SendScanRequest(msg *model.ScanRequestMessage) error {
	return x.sendSQSMessage(x.ScanQueueURL, msg)
}

func (x *Controller) SendInspectRequest(msg *model.InspectRequestMessage) error {
	return x.sendSQSMessage(x.InspectQueueURL, msg)
}
