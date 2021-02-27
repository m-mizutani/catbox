package main

import (
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		ctrl := controller.New()

		records, err := event.DecapSQSBody()
		if err != nil {
			return nil, err
		}

		for _, record := range records {
			var req model.ScanRequestMessage
			if err := record.Bind(&req); err != nil {
				return nil, err
			}

			if err := scanImage(ctrl, &req); err != nil {
				return nil, err
			}
		}

		return nil, nil
	})
}

func scanImage(ctrl *controller.Controller, req *model.ScanRequestMessage) error {
	trivyResults, err := ctrl.ScanImage(req.Target)
	if err != nil {
		return err
	}

	logger.With("report", trivyResults).With("req", req).Info("Scanned")
	return nil
}
