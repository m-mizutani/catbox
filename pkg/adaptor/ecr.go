package adaptor

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/pkg/errors"
)

// ECRPushImageEvent is PushImage event from CloudWatch Event
type ECRPushImageEvent struct {
	Account string `json:"account"`
	Detail  struct {
		AwsRegion         string `json:"awsRegion"`
		ErrorCode         string `json:"errorCode"`
		ErrorMessage      string `json:"errorMessage"`
		EventID           string `json:"eventID"`
		EventName         string `json:"eventName"`
		EventSource       string `json:"eventSource"`
		EventTime         string `json:"eventTime"`
		EventType         string `json:"eventType"`
		EventVersion      string `json:"eventVersion"`
		RequestID         string `json:"requestID"`
		RequestParameters struct {
			ImageManifest          string `json:"imageManifest"`
			ImageManifestMediaType string `json:"imageManifestMediaType"`
			ImageTag               string `json:"imageTag"`
			RegistryID             string `json:"registryId"`
			RepositoryName         string `json:"repositoryName"`
		} `json:"requestParameters"`
		ResponseElements struct {
			Image struct {
				ImageID struct {
					ImageDigest string `json:"imageDigest"`
					ImageTag    string `json:"imageTag"`
				} `json:"imageId"`
				ImageManifest          string `json:"imageManifest"`
				ImageManifestMediaType string `json:"imageManifestMediaType"`
				RegistryID             string `json:"registryId"`
				RepositoryName         string `json:"repositoryName"`
			} `json:"image"`
		} `json:"responseElements"`
		SourceIPAddress string `json:"sourceIPAddress"`
		UserAgent       string `json:"userAgent"`
	} `json:"detail"`
	DetailType string        `json:"detail-type"`
	ID         string        `json:"id"`
	Region     string        `json:"region"`
	Resources  []interface{} `json:"resources"`
	Source     string        `json:"source"`
	Time       string        `json:"time"`
	Version    string        `json:"version"`
}

// Layers returns image layer digests in ECRPushImageEvent.Detail.RequestParameters.ImageManifest
func (x ECRPushImageEvent) Layers() ([]string, error) {
	var manifest struct {
		Layers []struct {
			Digest string `json:"digest"`
		} `json:"layers"`
	}

	body := x.Detail.RequestParameters.ImageManifest
	if err := json.Unmarshal([]byte(body), &manifest); err != nil {
		// internal.Logger.WithField("body", body).WithError(err).Error("Fail json.Unmarshal")
		return nil, errors.Wrap(err, "Fail to parse ECRPushImageEvent.Detail.RequestParameters.ImageManifest")
	}

	var layers []string
	for _, layer := range manifest.Layers {
		layers = append(layers, layer.Digest)
	}

	return layers, nil
}

// ECRClient is ECR interface for catbox
type ECRClient interface {
	DescribeImagesPages(*ecr.DescribeImagesInput, func(*ecr.DescribeImagesOutput, bool) bool) error
	DescribeRepositoriesPages(*ecr.DescribeRepositoriesInput, func(*ecr.DescribeRepositoriesOutput, bool) bool) error
	GetAuthorizationToken(*ecr.GetAuthorizationTokenInput) (*ecr.GetAuthorizationTokenOutput, error)
}

// ECRClientFactory is interface for NewECRClient
type ECRClientFactory func(region string) (ECRClient, error)

// NewECRClient creates actual AWS SQS client, not stub
func NewECRClient(region string) (ECRClient, error) {
	ssn, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}
	return ecr.New(ssn), nil
}
