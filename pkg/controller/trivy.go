package controller

import (
	"path/filepath"
	"strings"
	"time"

	"encoding/json"

	"github.com/aquasecurity/trivy/pkg/report"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

// S3Loc describes location of S3 like directory
type S3Loc struct {
	Region string
	Bucket string
	Prefix string
}

type dbMetadata struct {
	MetadataVersion int `json:"Version"`
	// MetadataType is actually db.Type type
	MetadataType int       `json:"Type"`
	NextUpdate   time.Time `json:"NextUpdate"`
	UpdatedAt    string    `json:"UpdatedAt"`
}

const (
	trivyCacheDBKeyPart        = "cache/trivy/" + trivyCacheDBPathPart
	trivyCacheMetadataKeyPart  = "cache/trivy/" + trivyCacheMetadataPathPart
	trivyCacheDBPathPart       = "db/trivy.db"
	trivyCacheMetadataPathPart = "db/metadata.json"
)

func trivyCacheDBKey(prefix string) string {
	return prefix + trivyCacheDBKeyPart
}

func trivyCacheMetadataKey(prefix string) string {
	return prefix + trivyCacheMetadataKeyPart
}

func (x *Controller) downloadTrivyDB(localDir string) error {
	dbPath := filepath.Join(localDir, trivyCacheDBPathPart)

	if err := x.downloadS3Object(trivyCacheDBKeyPart, dbPath); err != nil {
		return golambda.WrapError(err).With("dbPath", dbPath)
	}

	metaPath := filepath.Join(localDir, trivyCacheMetadataPathPart)
	if err := x.downloadS3Object(trivyCacheMetadataKeyPart, metaPath); err != nil {
		return golambda.WrapError(err).With("metaPath", metaPath)
	}

	return nil
}

func (x *Controller) uploadTrivyReport(report *model.TrivyResults, req *model.ScanRequestMessage) error {
	return nil
}

func (x *Controller) DownloadTrivyReport(report *model.TrivyResults) {}

// ScanImage setup DB and metadata, then invokes trivy command by exec.Command
func (x *Controller) ScanImage(img model.Image) ([]report.Result, error) {
	localDir := "/tmp"

	if err := x.downloadTrivyDB(localDir); err != nil {
		return nil, err
	}

	imagePath := img.RegistryRepoTag()

	tmpName, err := x.TempFile("", "output*.json")
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to create temp file for trivy output")
	}

	trivyOptions := []string{
		"-q",
		"--cache-dir", localDir,
		"image",
		"--format", "json",
		"--skip-update",
		"-o", tmpName,
		// "--clear-cache",
		imagePath,
	}

	logger.
		With("localDir", localDir).
		With("img", img).
		With("options", trivyOptions).
		Info("Invoke trivy")

	ignoreErrors := []string{
		"unsupported MediaType: ",   // Ignore old container image schema
		"Invalid yarn.lock format:", // Ignore old yarn schema
	}

	out, err := x.Exec("./trivy", trivyOptions...)
	logger.With("out", string(out)).Debug("Done trivy command")

	if err != nil {
		for _, errmsg := range ignoreErrors {
			if strings.Index(string(out), errmsg) >= 0 {
				return nil, nil
			}
		}

		return nil, golambda.WrapError(err, "Fail to invoke trivy command")
	}

	trivyOutput, err := x.ReadFile(tmpName)
	if err != nil {
		return nil, golambda.WrapError(err, "Fail to read trivy output temp file").With("tmpName", tmpName)
	}

	results := []report.Result{}
	if err := json.Unmarshal(trivyOutput, &results); err != nil {
		return nil, golambda.WrapError(err, "Cannot Unmarshal Trivy output").With("out", string(trivyOutput))
	}

	return results, nil
}
