package main

import (
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		ctrl := controller.New()

		var cwEvent cloudWatchEvent
		if err := event.Bind(&cwEvent); err != nil {
			return nil, golambda.WrapError(err).With("event", event)
		}

		switch cwEvent.Source {
		case "aws.ecr":
			if err := handleECREvent(ctrl, cwEvent); err != nil {
				return nil, golambda.WrapError(err).With("event", cwEvent)
			}

		case "aws.events":
			if err := handlePeriodicEvent(ctrl, cwEvent); err != nil {
				return nil, golambda.WrapError(err).With("event", cwEvent)
			}

		default:
			logger.With("event", event).Info("WARNING: event source is not matched")
		}

		return nil, nil
	})
}
