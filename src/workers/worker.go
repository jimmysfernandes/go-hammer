package workers

import "context"

type WorkResult struct {
	ExpectedIterations uint16 `json:"expected_iterations"`
	ExpectedTasks      uint16 `json:"expected_tasks"`
	ExecutedIterations uint16 `json:"executed_iterations"`
	ExecutedTasks      uint16 `json:"executed_tasks"`
	ErrorTasks         uint16 `json:"error_tasks"`
	ErrorInterations   uint16 `json:"error_iterations"`
}

type Worker interface {
	DoWork(ctx context.Context, task func(ctx context.Context) error) (WorkResult, error)
}
