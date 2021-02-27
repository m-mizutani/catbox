package main

import (
	"github.com/m-mizutani/catbox/pkg/interfaces"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/catbox/pkg/service"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		config, err := interfaces.NewConfig()
		if err != nil {
			return nil, golambda.WrapError(err).With("event", event)
		}

		records, err := event.DecapSQSBody()
		if err != nil {
			return nil, err
		}

		for _, record := range records {
			var req model.ScanRequestMessage
			if err := record.Bind(&req); err != nil {
				return nil, err
			}

			if err := scanImage(config, &req); err != nil {
				return nil, err
			}
		}

		return nil, nil
	})
}

func scanImage(config *interfaces.Config, req *model.ScanRequestMessage) error {
	svc := service.New(config)
	trivyResults, err := svc.ScanImage(req.Target)
	if err != nil {
		return err
	}

	logger.With("report", trivyResults).With("req", req).Info("Scanned")
	return nil
}
