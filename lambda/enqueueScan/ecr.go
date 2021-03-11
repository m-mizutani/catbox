package main

import (
	"fmt"
	"strings"

	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/catbox/pkg/usecase"
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

	for _, rsc := range event.Detail.Resources {
		registry, repo, err := parseRepositoryARN(rsc.Arn)
		if err != nil {
			return err
		}

		target := &model.Image{
			Registry: registry,
			Repo:     repo,
			Tag:      event.Detail.ResponseElements.Image.ImageID.ImageTag,
			Digest:   event.Detail.ResponseElements.Image.ImageID.ImageDigest,
		}

		if err := usecase.EnqueueScanRequest(ctrl, target, "ecr.PutImage"); err != nil {
			return err
		}
	}

	return nil
}
