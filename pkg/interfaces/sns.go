package interfaces

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type SNSClient interface {
	Publish(input *sns.PublishInput) (*sns.PublishOutput, error)
}

type SNSClientFactory func(region string) (SNSClient, error)

func NewSNSClient(region string) (SNSClient, error) {
	ssn, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}
	return sns.New(ssn), nil

}
