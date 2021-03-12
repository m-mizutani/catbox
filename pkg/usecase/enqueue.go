package usecase

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

// EnqueueScanRequest retrieves layer digests and environment variables of a target image and sends them to SQS as scan request message.
func EnqueueScanRequest(ctrl *controller.Controller, target *model.TaggedImage, requester string) error {
	reqID := uuid.New().String()
	ts := time.Now().UTC()
	basePrefix := fmt.Sprintf("snapshots/%s/%s/%s_%s/", target.Registry, target.Repo, ts.Format("20060102_150405"), reqID)

	req := &model.ScanRequestMessage{
		RequestID:   reqID,
		RequestedBy: requester,
		RequestedAt: ts,
		Target:      *target,
		OutS3Prefix: basePrefix,
	}
	if err := ctrl.SendScanRequest(req); err != nil {
		return golambda.WrapError(err).With("req", req)
	}
	logger.With("req", req).Info("Sent scan request")

	return nil
}

// GetImageMetaData retrieves image layer digests and environment variables of image. It can specify only tagged image (registry, repository and tag) because trivy supports only tag based scan for remote image.
func GetImageMetaData(ctrl *controller.Controller, target *model.TaggedImage) (*model.ImageMeta, error) {
	token, err := ctrl.GetRegistryAPIToken(target.Registry)
	if err != nil {
		return nil, golambda.WrapError(err).With("target", target)
	}

	var meta model.ImageMeta

	// Fill layer digests
	manifest, err := ctrl.GetImageManifest(&model.Image{
		Registry: target.Registry,
		Repo:     target.Repo,
	}, target.Tag, token)
	if err != nil {
		return nil, golambda.WrapError(err).With("target", target)
	}
	meta.LayerDigests = make([]string, len(manifest.Layers))
	for i := range manifest.Layers {
		meta.LayerDigests[i] = manifest.Layers[i].Digest
	}

	// Fill env vars
	envVars, err := ctrl.GetImageEnv(manifest, target, token)
	if err != nil {
		return nil, golambda.WrapError(err).With("manifest", manifest).With("target", target)
	}
	meta.Env = envVars

	return &meta, nil
}
