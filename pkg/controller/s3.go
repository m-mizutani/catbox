package controller

import (
	"bytes"
	"io"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

// downloadS3Object downloads object s3://{x.S3Bucket}/{x.S3Prefix}{suffixKey} to file:///{dstPath}
func (x *Controller) downloadS3ToFile(suffixKey, dstPath string) error {
	client, err := x.adaptors.NewS3(x.S3Region)
	if err != nil {
		return err
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(x.S3Bucket),
		Key:    aws.String(x.S3Prefix + suffixKey),
	}

	output, err := client.GetObject(input)
	if err != nil {
		return golambda.WrapError(err, "Fail to download DB").With("input", input)
	}
	defer output.Body.Close()

	if err := x.adaptors.MkdirAll(filepath.Dir(dstPath), 0777); err != nil {
		return golambda.WrapError(err, "Failed to create %s directory", filepath.Dir(dstPath))
	}

	file, err := x.adaptors.Create(dstPath)
	if err != nil {
		return golambda.WrapError(err, "cannot save file")
	}
	defer file.Close()

	if _, err := io.Copy(file, output.Body); err != nil {
		return golambda.WrapError(err, "Fail to copy S3 data to local file")
	}
	return nil
}

func (x *Controller) uploadFileToS3(suffixKey, srcPath string) error {
	data, err := x.adaptors.ReadFile(srcPath)
	if err != nil {
		return golambda.WrapError(err, "Read file to upload S3 object").With("srcPath", srcPath)
	}

	if _, err := x.uploadS3Data(suffixKey, bytes.NewReader(data), nil); err != nil {
		return err
	}
	return nil
}

func (x *Controller) uploadS3Data(suffixKey string, reader io.ReadSeeker, encoding *string) (*model.S3Path, error) {
	client, err := x.adaptors.NewS3(x.S3Region)
	if err != nil {
		return nil, err
	}

	s3Key := x.S3Prefix + suffixKey
	input := &s3.PutObjectInput{
		Bucket:          aws.String(x.S3Bucket),
		Key:             aws.String(s3Key),
		Body:            reader,
		ContentEncoding: encoding,
	}

	if _, err := client.PutObject(input); err != nil {
		return nil, golambda.WrapError(err, "PutObject").With("input", input)
	}
	return &model.S3Path{
		Region: x.S3Region,
		Bucket: x.S3Bucket,
		Key:    s3Key,
	}, nil
}
