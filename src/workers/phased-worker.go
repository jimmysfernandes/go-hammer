package workers

import (
	"context"
	"sync"
	"time"

	"github.com/jimmysfernandes/go-hammer/src/metrics"
)

type SequenceWorker struct {
	Itrations         uint16 `json:"itrations"`
	TasksPerIteration uint16 `json:"tasks_per_iteration"`
}

type PhasedWorker struct {
	wg               sync.WaitGroup
	sequences        []SequenceWorker
	errorInterations metrics.Counter
	executedTasks    metrics.Counter
	errorTasks       metrics.Counter
}

func (c *PhasedWorker) DoWork(ctx context.Context, task func(ctx context.Context) error) (WorkResult, error) {
	// Reset counters
	defer c.resetState()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	stop := false

	for i := 0; i < len(c.sequences) && !stop; i++ {
		sequence := c.sequences[i]

		select {
		case <-ctx.Done():
			stop = true
		default:
			c.wg.Add(1)
			go func(s SequenceWorker) {
				defer c.wg.Done()
				c.runSequence(ctx, task, s)
			}(sequence)

			time.Sleep(time.Duration(sequence.Itrations) * time.Second)
		}
	}

	c.wg.Wait()

	return WorkResult{
		ExpectedIterations: c.expectedIterations(),
		ExpectedTasks:      c.expectedTasks(),
		ExecutedIterations: uint16(len(c.sequences)),
		ExecutedTasks:      uint16(c.executedTasks.Value()),
		ErrorInterations:   uint16(c.errorInterations.Value()),
		ErrorTasks:         uint16(c.errorTasks.Value()),
	}, nil
}

func (c *PhasedWorker) expectedIterations() uint16 {
	expectedIterations := uint16(0)
	for _, sequence := range c.sequences {
		expectedIterations += sequence.Itrations
	}
	return expectedIterations
}

func (c *PhasedWorker) expectedTasks() uint16 {
	expectedTasks := uint16(0)
	for _, sequence := range c.sequences {
		expectedTasks += sequence.Itrations * sequence.TasksPerIteration
	}
	return expectedTasks
}

func (c *PhasedWorker) resetState() {
	c.errorInterations.Reset()
	c.executedTasks.Reset()
	c.errorTasks.Reset()
}

func (c *PhasedWorker) runSequence(ctx context.Context, task func(ctx context.Context) error, sequence SequenceWorker) error {
	select {
	case <-ctx.Done():
		return nil
	default:
		c.wg.Add(1)
		constatWorker := NewConstantWorker(sequence.Itrations, sequence.TasksPerIteration)
		go func() {
			defer c.wg.Done()
			constatWorker.DoWork(ctx, task)
		}()
	}

	return nil
}

func NewPhasedWorker(sequences []SequenceWorker) Worker {
	return &PhasedWorker{
		wg:        sync.WaitGroup{},
		sequences: sequences,
	}
}
