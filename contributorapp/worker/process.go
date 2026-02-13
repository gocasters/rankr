package worker

import (
	"context"
	"fmt"
	"strconv"
)

type Processor interface {
	ProcessJob(ctx context.Context, jobID uint) error
}

type Worker struct {
	processor Processor
}

func NewProcess(p Processor) Worker {
	return Worker{processor: p}
}

func (p Worker) Process(ctx context.Context, jobID string) error {
	if jobID == "" {
		return fmt.Errorf("job id is empty")
	}

	uID, err := strconv.ParseUint(jobID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid job id %q: %w", jobID, err)
	}

	err = p.processor.ProcessJob(ctx, uint(uID))
	if err != nil {
		return err
	}

	return nil
}
