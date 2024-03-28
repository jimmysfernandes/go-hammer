package workers

import (
	"context"
	"sync"
	"time"

	"github.com/jimmysfernandes/go-hammer/src/metrics"
)

type ConstantWorker struct {
	wg                sync.WaitGroup
	itrations         uint16
	tasksPerIteration uint16
	errorInterations  metrics.Counter
	executedTasks     metrics.Counter
	errorTasks        metrics.Counter
}

func (c *ConstantWorker) DoWork(ctx context.Context, task func(ctx context.Context) error) (WorkResult, error) {
	// Reset counters
	defer c.resetState()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	stop := false
	executedIterations := uint16(1)

	for !stop && executedIterations <= c.itrations {
		select {
		case <-ctx.Done():
			stop = true
		case <-ticker.C:
			c.wg.Add(1)
			go func() {
				defer c.wg.Done()
				if err := c.runIteration(ctx, task); err != nil {
					c.errorInterations.Inc()
				}
			}()
			executedIterations++
		}
	}

	c.wg.Wait()

	return WorkResult{
		ExpectedIterations: c.itrations,
		ExpectedTasks:      c.itrations * c.tasksPerIteration,
		ExecutedIterations: executedIterations,
		ExecutedTasks:      uint16(c.executedTasks.Value()),
		ErrorInterations:   uint16(c.errorInterations.Value()),
		ErrorTasks:         uint16(c.errorTasks.Value()),
	}, nil
}

func (c *ConstantWorker) resetState() {
	c.errorInterations.Reset()
	c.executedTasks.Reset()
	c.errorTasks.Reset()
}

func (c *ConstantWorker) runIteration(ctx context.Context, task func(ctx context.Context) error) error {
	for i := uint16(0); i < c.tasksPerIteration; i++ {
		select {
		case <-ctx.Done():
			return nil
		default:
			c.wg.Add(1)
			go func() {
				defer c.wg.Done()
				defer c.executedTasks.Inc()
				if err := task(ctx); err != nil {
					c.errorTasks.Inc()
				}
			}()
		}
	}

	return nil
}

func NewConstantWorker(itrations uint16, tasksPerIteration uint16) *ConstantWorker {
	return &ConstantWorker{
		wg:                sync.WaitGroup{},
		itrations:         itrations,
		tasksPerIteration: tasksPerIteration,
		errorInterations:  metrics.NewCounter(),
		executedTasks:     metrics.NewCounter(),
		errorTasks:        metrics.NewCounter(),
	}
}
