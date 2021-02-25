package main

import (
	"github.com/m-mizutani/catbox/pkg/handler"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		args, err := handler.NewArguments()
		if err != nil {
			return nil, golambda.WrapError(err).With("event", event)
		}

		var cwEvent cloudWatchEvent
		if err := event.Bind(&cwEvent); err != nil {
			return nil, golambda.WrapError(err).With("event", event)
		}

		switch cwEvent.Source {
		case "aws.ecr":
			if err := handleECREvent(args, cwEvent); err != nil {
				return nil, golambda.WrapError(err).With("event", cwEvent)
			}

		case "aws.events":
			if err := handlePeriodicEvent(args, cwEvent); err != nil {
				return nil, golambda.WrapError(err).With("event", cwEvent)
			}

		default:
			logger.With("event", event).Info("WARNING: event source is not matched")
		}

		return nil, nil
	})
}
