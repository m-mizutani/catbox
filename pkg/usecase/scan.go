package usecase

import (
	"time"

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

	invokedAt := time.Now().UTC()
	trivyResults, err := ctrl.InvokeTrivyScan(req.Target, cacheDir)
	if err != nil {
		return err
	}
	scannedAt := time.Now().UTC()

	logger.With("len(results)", len(trivyResults)).With("req", req).Info("Scanned")
	logger.With("results", trivyResults).Debug("Scanned results")

	outputPath, err := ctrl.UploadTrivyReport(trivyResults, req.S3Key(model.ScanTypeTrivy))
	if err != nil {
		return err
	}

	report := &model.ScanReport{
		Image:       req.Target,
		ImageMeta:   req.TargetMeta,
		ScanType:    model.ScanTypeTrivy,
		RequestedAt: req.RequestedAt.UTC().Unix(),
		RequestedBy: req.RequestedBy,
		InvokedAt:   invokedAt.Unix(),
		ScannedAt:   scannedAt.Unix(),
		OutputTo:    *outputPath,
	}
	if err := ctrl.DB().PutScanReport(report); err != nil {
		return err
	}

	logger.With("report", report).Info("Saved report")

	if err := ctrl.SendInspectRequest(&model.InspectRequestMessage{
		ReportID: report.ReportID,
	}); err != nil {
		return err
	}

	return nil
}
