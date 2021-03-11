package controller

import (
	"bytes"
	"compress/gzip"
	"path/filepath"
	"strings"
	"time"

	"encoding/json"

	"github.com/aquasecurity/trivy/pkg/report"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

type dbMetadata struct {
	MetadataVersion int `json:"Version"`
	// MetadataType is actually db.Type type
	MetadataType int       `json:"Type"`
	NextUpdate   time.Time `json:"NextUpdate"`
	UpdatedAt    string    `json:"UpdatedAt"`
}

const (
	trivyCacheDBPathPart       = "db/trivy.db"
	trivyCacheMetadataPathPart = "db/metadata.json"
	trivyCacheDBKeyPart        = "cache/trivy/" + trivyCacheDBPathPart
	trivyCacheMetadataKeyPart  = "cache/trivy/" + trivyCacheMetadataPathPart
)

func trivyCacheDBKey(prefix string) string {
	return prefix + trivyCacheDBKeyPart
}

func trivyCacheMetadataKey(prefix string) string {
	return prefix + trivyCacheMetadataKeyPart
}

func trivyCacheDBPath(cacheDir string) string {
	return filepath.Join(cacheDir, trivyCacheDBPathPart)
}

func trivyCacheMetadataPath(cacheDir string) string {
	return filepath.Join(cacheDir, trivyCacheMetadataPathPart)
}

// DownloadTrivyDB gets trivy DB and metadata.json and saves them to cacheDir
func (x *Controller) DownloadTrivyDB(cacheDir string) error {
	dbPath := trivyCacheDBPath(cacheDir)

	if err := x.downloadS3ToFile(trivyCacheDBKeyPart, dbPath); err != nil {
		return golambda.WrapError(err).With("dbPath", dbPath)
	}

	metaPath := trivyCacheMetadataPath(cacheDir)
	if err := x.downloadS3ToFile(trivyCacheMetadataKeyPart, metaPath); err != nil {
		return golambda.WrapError(err).With("metaPath", metaPath)
	}

	return nil
}

// UploadTrivyReport saves report to dst as S3 object
func (x *Controller) UploadTrivyReport(report model.TrivyResults, suffixKey string) (*model.S3Path, error) {
	raw, err := json.Marshal(report)
	if err != nil {
		return nil, golambda.WrapError(err, "json.Marshal trivy report").With("report", report)
	}

	buf := &bytes.Buffer{}
	wr := gzip.NewWriter(buf)
	if _, err := wr.Write(raw); err != nil {
		return nil, golambda.WrapError(err, "gzip.Write trivy report").With("raw", string(raw))
	}
	if err := wr.Close(); err != nil {
		return nil, golambda.WrapError(err, "gzip.Close trivy report")
	}

	s3Path, err := x.uploadS3Data(suffixKey, bytes.NewReader(buf.Bytes()), aws.String("gzip"))
	if err != nil {
		return nil, golambda.WrapError(err, "s3.PutObject results of trivy").With("suffixKey", suffixKey)
	}

	return s3Path, nil
}

// DownloadTrivyReport gets report from src S3 object. src indicates direct s3 path.
func (x *Controller) DownloadTrivyReport(src *model.S3Path) (model.TrivyResults, error) {
	s3Client, err := x.adaptors.NewS3(src.Region)
	if err != nil {
		return nil, err
	}

	input := &s3.GetObjectInput{
		Bucket: &src.Bucket,
		Key:    &src.Key,
	}
	output, err := s3Client.GetObject(input)
	if err != nil {
		return nil, golambda.WrapError(err, "s3.GetObject").With("input", input)
	}

	var results model.TrivyResults
	if err := json.NewDecoder(output.Body).Decode(&results); err != nil {
		return nil, golambda.WrapError(err, "json.Decoder.Decode").With("output", output)
	}

	return results, nil
}

// HasTrivyDB checks if both of DB and metadata.json exist
func (x *Controller) HasTrivyDB(cacheDir string) bool {

	return false
}

// InvokeTrivyScan setup DB invokes trivy command by exec.Command
func (x *Controller) InvokeTrivyScan(img model.TaggedImage, cacheDir string) ([]report.Result, error) {
	imagePath := img.RegistryRepoTag()

	tmpName, err := x.adaptors.TempFile("", "output*.json")
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to create temp file for trivy output")
	}

	trivyOptions := []string{
		"-q",
		"--cache-dir", cacheDir,
		"image",
		"--format", "json",
		"--skip-update",
		"-o", tmpName,
		// "--clear-cache",
		imagePath,
	}

	logger.
		With("cacheDir", cacheDir).
		With("img", img).
		With("options", trivyOptions).
		Info("Invoke trivy")

	ignoreErrors := []string{
		"unsupported MediaType: ",   // Ignore old container image schema
		"Invalid yarn.lock format:", // Ignore old yarn schema
	}

	out, err := x.adaptors.Exec("./trivy", trivyOptions...)
	logger.With("out", string(out)).Debug("Done trivy command")

	if err != nil {
		for _, errmsg := range ignoreErrors {
			if strings.Index(string(out), errmsg) >= 0 {
				return nil, nil
			}
		}

		return nil, golambda.WrapError(err, "Fail to invoke trivy command")
	}

	trivyOutput, err := x.adaptors.ReadFile(tmpName)
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to read trivy output temp file").With("tmpName", tmpName)
	}

	results := []report.Result{}
	if err := json.Unmarshal(trivyOutput, &results); err != nil {
		return nil, golambda.WrapError(err, "Cannot Unmarshal Trivy output").With("out", string(trivyOutput))
	}

	return results, nil
}
