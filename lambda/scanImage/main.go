package main

import (
	"github.com/m-mizutani/catbox/pkg/handler"
	"github.com/m-mizutani/catbox/pkg/models"
	"github.com/m-mizutani/catbox/pkg/service"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		args, err := handler.NewArguments()
		if err != nil {
			return nil, golambda.WrapError(err).With("event", event)
		}

		records, err := event.DecapSQSBody()
		if err != nil {
			return nil, err
		}

		for _, record := range records {
			var req models.ScanRequestMessage
			if err := record.Bind(&req); err != nil {
				return nil, err
			}

			if err := scanImage(args, &req); err != nil {
				return nil, err
			}
		}

		return nil, nil
	})
}

func scanImage(args *handler.Arguments, req *models.ScanRequestMessage) error {
	svc := service.New(args)
	trivyResults, err := svc.ScanImage(req.Target)
	if err != nil {
		return err
	}

	logger.With("report", trivyResults).With("req", req).Info("Scanned")
	return nil
}
