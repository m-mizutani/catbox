package usecase_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/m-mizutani/catbox/pkg/interfaces"
)

type (
	mockS3 struct {
		regions        map[string]struct{}
		getObjectInput []*s3.GetObjectInput
		putObjectInput []*s3.PutObjectInput

		objects map[string]map[string]*s3.GetObjectOutput
	}
	mockSQS struct {
		input []*sqs.SendMessageInput
	}
	mockSNS struct {
		input []*sns.PublishInput
	}
	mockECR struct {
		getTokenInput  []*ecr.GetAuthorizationTokenInput
		getTokenOutput []*ecr.GetAuthorizationTokenOutput
		interfaces.ECRClient
	}
	bufCloser struct{ bytes.Buffer }
	mockHTTP  struct {
		requests  []*http.Request
		responses []*http.Response
	}
	mockSet struct {
		s3       mockS3
		sqs      mockSQS
		sns      mockSNS
		ecr      mockECR
		http     mockHTTP
		buffers  []*bufCloser
		dbClient interfaces.DBClient
	}
)

func (x *mockS3) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	x.getObjectInput = append(x.getObjectInput, input)

	bucket, ok := x.objects[*input.Bucket]
	if !ok {
		return nil, errors.New(s3.ErrCodeNoSuchBucket)
	}
	out, ok := bucket[*input.Key]
	if !ok {
		return nil, errors.New(s3.ErrCodeNoSuchKey)
	}
	return out, nil
}

func (x *mockS3) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	x.putObjectInput = append(x.putObjectInput, input)
	return &s3.PutObjectOutput{}, nil
}

func (x *mockS3) saveObject(bucket, key string, obj interface{}) {
	if x.objects == nil {
		x.objects = make(map[string]map[string]*s3.GetObjectOutput)
	}

	var reader io.Reader
	switch v := obj.(type) {
	case string:
		reader = strings.NewReader(v)
	case []byte:
		reader = bytes.NewReader(v)
	default:
		raw, err := json.Marshal(obj)
		if err != nil {
			panic("marshal object: " + err.Error())
		}
		reader = bytes.NewReader(raw)
	}

	bkt, ok := x.objects[bucket]
	if !ok {
		bkt = make(map[string]*s3.GetObjectOutput)
		x.objects[bucket] = bkt
	}

	bkt[key] = &s3.GetObjectOutput{
		Body: ioutil.NopCloser(reader),
	}
}

func (x *mockSQS) SendMessage(input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	x.input = append(x.input, input)
	return &sqs.SendMessageOutput{}, nil
}

func (x *mockSNS) Publish(input *sns.PublishInput) (*sns.PublishOutput, error) {
	x.input = append(x.input, input)
	return &sns.PublishOutput{}, nil
}

func (x *mockECR) GetAuthorizationToken(input *ecr.GetAuthorizationTokenInput) (*ecr.GetAuthorizationTokenOutput, error) {
	output := x.getTokenOutput[len(x.getTokenInput)]
	x.getTokenInput = append(x.getTokenInput, input)
	return output, nil
}

func (x *bufCloser) Close() error { return nil }

func (x *mockHTTP) Do(req *http.Request) (*http.Response, error) {
	resp := x.responses[len(x.requests)]
	x.requests = append(x.requests, req)
	return resp, nil
}
