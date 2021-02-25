package main

import "github.com/m-mizutani/golambda"

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		return nil, nil
	})
}
