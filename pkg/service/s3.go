package service

import (
	"bytes"
	"io"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/m-mizutani/golambda"
)

// Hint: Not thread safe
func (x *Service) setupS3Client() error {
	if x.s3Client != nil {
		return nil
	}

	// catbox accesses to S3 bucket specified by S3Region.
	s3Client, err := x.args.Adaptors.NewS3(x.args.S3Region)
	if err != nil {
		return golambda.WrapError(err, "Creating S3 client").With("region", x.args.S3Region)
	}

	x.s3Client = s3Client
	return nil
}

// downloadS3Object downloads object s3://{x.args.S3Bucket}/{x.args.S3Prefix}{suffixKey} to file:///{dstPath}
func (x *Service) downloadS3Object(suffixKey, dstPath string) error {
	if err := x.setupS3Client(); err != nil {
		return err
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(x.args.S3Bucket),
		Key:    aws.String(x.args.S3Prefix + suffixKey),
	}

	output, err := x.s3Client.GetObject(input)
	if err != nil {
		return golambda.WrapError(err, "Fail to download DB").With("input", input)
	}
	defer output.Body.Close()

	if err := x.args.Adaptors.MkdirAll(filepath.Dir(dstPath), 0777); err != nil {
		return golambda.WrapError(err, "Failed to create %s directory", filepath.Dir(dstPath))
	}

	file, err := x.args.Adaptors.Create(dstPath)
	if err != nil {
		return golambda.WrapError(err, "cannot save file")
	}
	defer file.Close()

	if _, err := io.Copy(file, output.Body); err != nil {
		return golambda.WrapError(err, "Fail to copy S3 data to local file")
	}
	return nil
}

func (x *Service) uploadS3Object(suffixKey, srcPath string) error {
	if err := x.setupS3Client(); err != nil {
		return err
	}

	data, err := x.args.Adaptors.ReadFile(srcPath)
	if err != nil {
		return golambda.WrapError(err, "Read file to upload S3 object").With("srcPath", srcPath)
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(x.args.S3Bucket),
		Key:    aws.String(x.args.S3Prefix + suffixKey),
		Body:   bytes.NewReader(data),
	}

	if _, err := x.s3Client.PutObject(input); err != nil {
		return golambda.WrapError(err, "PutObject").With("input", input)
	}

	return nil
}
