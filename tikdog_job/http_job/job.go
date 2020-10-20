package http_job

import (
	"context"
)

type job struct {
	name    string
	version string
	kind    string
	code    string
}

func (t *job) OnEvent(ctx context.Context, event interface{}) error {
	if event == nil {
		return nil
	}

	switch event := event.(type) {
	default:
		return nil
	}
}
