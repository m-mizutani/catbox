package controller

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

func (x *Controller) sendSQSMessage(url string, data interface{}) error {
	// SQS region must be same with Lambda region
	client, err := x.adaptors.NewSQS(x.AwsRegion)
	if err != nil {
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
	if _, err := client.SendMessage(input); err != nil {
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
