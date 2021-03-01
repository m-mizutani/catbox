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
func EnqueueScanRequest(ctrl *controller.Controller, target *model.Image, requester string) error {
	token, err := ctrl.GetRegistryAPIToken(target.Registry)
	if err != nil {
		return golambda.WrapError(err).With("target", target)
	}

	var meta model.ImageMeta

	// Fill layer digests
	manifest, err := ctrl.GetImageManifest(target, token)
	if err != nil {
		return golambda.WrapError(err).With("target", target)
	}
	meta.LayerDigests = make([]string, len(manifest.Layers))
	for i := range manifest.Layers {
		meta.LayerDigests[i] = manifest.Layers[i].Digest
	}

	// Fill env vars
	envVars, err := ctrl.GetImageEnv(manifest, target, token)
	if err != nil {
		return golambda.WrapError(err).With("manifest", manifest).With("target", target)
	}
	meta.Env = envVars

	ts := time.Now().UTC()
	basePrefix := fmt.Sprintf("snapshots/%s/%s/%s/%s_%06d/", target.Registry, target.Repo, target.Digest, ts.Format("20060102_150405"), ts.Nanosecond()/1000)

	req := &model.ScanRequestMessage{
		RequestID:   uuid.New().String(),
		RequestedBy: requester,
		RequestedAt: ts,
		Target:      *target,
		TargetMeta:  meta,
		OutS3Prefix: basePrefix,
	}
	if err := ctrl.SendScanRequest(req); err != nil {
		return golambda.WrapError(err).With("req", req)
	}
	logger.With("req", req).Info("Sent scan request")

	return nil
}
