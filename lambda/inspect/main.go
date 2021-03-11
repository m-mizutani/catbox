package main

import (
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/catbox/pkg/model"
	"github.com/m-mizutani/catbox/pkg/usecase"
	"github.com/m-mizutani/golambda"
)

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		ctrl := controller.New()

		records, err := event.DecapSQSBody()
		if err != nil {
			return nil, err
		}

		for _, record := range records {
			var req model.InspectRequestMessage
			if err := record.Bind(&req); err != nil {
				return nil, err
			}

			if err := usecase.InspectScanReport(ctrl, &req); err != nil {
				return nil, err
			}
		}

		return nil, nil
	})
}
