package controller

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

func extractRegionFromTopicARN(arn string) (string, error) {
	// Example) arn:aws:sns:us-east-1:111122223333:my-topic
	// [0]: arn
	// [1]: aws
	// [2]: sns
	// [3]: us-east-1
	// [4]: 111122223333
	// [5]: my-topic
	parts := strings.Split(arn, ":")
	if len(parts) != 6 {
		return "", golambda.NewError("Invalid Topic ARN").With("arn", arn)
	}

	return parts[3], nil
}

func (x *Controller) publishSNS(arn string, obj interface{}) error {
	region, err := extractRegionFromTopicARN(x.ChangeTopicARN)
	if err != nil {
		return err
	}

	client, err := x.adaptors.NewSNS(region)
	if err != nil {
		return golambda.WrapError(err, "NewSNS").With("region", region)
	}

	raw, err := json.Marshal(obj)
	if err != nil {
		return golambda.WrapError(err, "json.Marshal changeMessage").With("obj", obj)
	}
	input := &sns.PublishInput{
		TopicArn: aws.String(arn),
		Message:  aws.String(string(raw)),
	}

	if _, err := client.Publish(input); err != nil {
		return golambda.WrapError(err, "sns.Publish").With("input", input)
	}

	return nil
}

// PublishChangeMessage sends message of json.Marshal model.ChangeMessage
func (x *Controller) PublishChangeMessage(msg *model.ChangeMessage) error {
	return x.publishSNS(x.ChangeTopicARN, msg)
}
