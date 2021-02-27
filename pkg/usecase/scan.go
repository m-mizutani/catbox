package usecase

import (
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/catbox/pkg/model"
)

// TrivyScanImage setup trivy DB and scan image, then upload results to S3
func TrivyScanImage(ctrl *controller.Controller, req *model.ScanRequestMessage) error {
	cacheDir := "/tmp"
	if !ctrl.HasTrivyDB(cacheDir) {
		if err := ctrl.DownloadTrivyDB(cacheDir); err != nil {
			return err
		}
	}

	trivyResults, err := ctrl.InvokeTrivyScan(req.Target, cacheDir)
	if err != nil {
		return err
	}

	logger.With("report", trivyResults).With("req", req).Info("Scanned")

	ctrl.UploadTrivyReport(trivyResults, &model.S3Path{
		Region: req.S3Bucket,
		Bucket: req.S3Bucket,
		Key:    req.S3Key(model.ScanTypeTrivy),
	})
	return nil
}
