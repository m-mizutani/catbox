package controller

import (
	"bytes"
	"io"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/m-mizutani/golambda"
)

// Hint: Not thread safe
func (x *Controller) setupS3Client() error {
	if x.s3Client != nil {
		return nil
	}

	// catbox accesses to S3 bucket specified by S3Region.
	s3Client, err := x.NewS3(x.S3Region)
	if err != nil {
		return golambda.WrapError(err, "Creating S3 client").With("region", x.S3Region)
	}

	x.s3Client = s3Client
	return nil
}

// downloadS3Object downloads object s3://{x.S3Bucket}/{x.S3Prefix}{suffixKey} to file:///{dstPath}
func (x *Controller) downloadS3Object(suffixKey, dstPath string) error {
	if err := x.setupS3Client(); err != nil {
		return err
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(x.S3Bucket),
		Key:    aws.String(x.S3Prefix + suffixKey),
	}

	output, err := x.s3Client.GetObject(input)
	if err != nil {
		return golambda.WrapError(err, "Fail to download DB").With("input", input)
	}
	defer output.Body.Close()

	if err := x.MkdirAll(filepath.Dir(dstPath), 0777); err != nil {
		return golambda.WrapError(err, "Failed to create %s directory", filepath.Dir(dstPath))
	}

	file, err := x.Create(dstPath)
	if err != nil {
		return golambda.WrapError(err, "cannot save file")
	}
	defer file.Close()

	if _, err := io.Copy(file, output.Body); err != nil {
		return golambda.WrapError(err, "Fail to copy S3 data to local file")
	}
	return nil
}

func (x *Controller) uploadS3Object(suffixKey, srcPath string) error {
	if err := x.setupS3Client(); err != nil {
		return err
	}

	data, err := x.ReadFile(srcPath)
	if err != nil {
		return golambda.WrapError(err, "Read file to upload S3 object").With("srcPath", srcPath)
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(x.S3Bucket),
		Key:    aws.String(x.S3Prefix + suffixKey),
		Body:   bytes.NewReader(data),
	}

	if _, err := x.s3Client.PutObject(input); err != nil {
		return golambda.WrapError(err, "PutObject").With("input", input)
	}

	return nil
}
