package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

func parseRepositoryARN(arn string) (string, string, error) {
	// Example) arn:aws:ecr:ap-northeast-1:1111111111:repository/my-awesome-app
	// [0] arn
	// [1] aws
	// [2] ecr
	// [3] ap-northeast-1
	// [4] 1111111111
	// [5] repository/my-awesome-app

	arnParts := strings.Split(arn, ":")
	if len(arnParts) != 6 {
		return "", "", golambda.NewError("Failed to parse Repository ARN (colon num is not matched)").With("arn", arn)
	}

	repoParts := strings.Split(arnParts[5], "/")
	if len(repoParts) != 2 {
		return "", "", golambda.NewError("Failed to parse Repository ARN (slash num is not matched)").With("arn", arn)
	}

	return fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", arnParts[4], arnParts[3]), repoParts[1], nil
}

func handleECREvent(ctrl *controller.Controller, event cloudWatchEvent) error {
	logger.With("event", event).Info("handleECREvent")

	if "PutImage" != event.Detail.EventName {
		logger.With("event", event).Info("WARNING: ECR event is not PutImage")
		return nil
	}

	ts, err := time.Parse("2006-01-02T15:04:05Z", event.Detail.EventTime)
	if err != nil {
		return golambda.WrapError(err, "Parsing eventTime of ECR event")
	}

	for _, rsc := range event.Detail.Resources {
		registry, repo, err := parseRepositoryARN(rsc.Arn)
		if err != nil {
			return err
		}

		target := model.Image{
			Registry: registry,
			Repo:     repo,
			Tag:      event.Detail.ResponseElements.Image.ImageID.ImageTag,
			Digest:   event.Detail.ResponseElements.Image.ImageID.ImageDigest,
		}
		basePrefix := fmt.Sprintf("snapshots/%s/%s/%s/%s_%06d/", target.Registry, target.Repo, target.Digest, ts.Format("20060102_150405"), ts.Nanosecond()/1000)

		req := &model.ScanRequestMessage{
			RequestID:   uuid.New().String(),
			RequestedBy: "ecr.PushImage",
			RequestedAt: ts,
			Target:      target,
			OutS3Prefix: basePrefix,
		}

		if err := ctrl.SendScanRequest(req); err != nil {
			return golambda.WrapError(err).With("req", req)
		}
		logger.With("req", req).Info("Sent scan request")
	}

	return nil
}
