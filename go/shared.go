package app

import (
	"go.temporal.io/sdk/workflow"
)

const BreakfastTaskQueue = "BREAKFAST_TASK_QUEUE"

type Compensations []any

func (s *Compensations) AddCompensation(activity any) {
	*s = append(*s, activity)
}

func (s Compensations) Compensate(ctx workflow.Context, inParallel bool) {
	if !inParallel {
		for i := len(s) - 1; i >= 0; i-- {
			errCompensation := workflow.ExecuteActivity(ctx, s[i]).Get(ctx, nil)
			if errCompensation != nil {
				workflow.GetLogger(ctx).Error("Executing compensation failed", errCompensation)
			}
		}
	} else {
		var pendingFutures []workflow.Future
		selector := workflow.NewSelector(ctx)
		for i := 0; i < len(s); i++ {
			workflow_execution := workflow.ExecuteActivity(ctx, s[i])
			pendingFutures = append(pendingFutures, workflow_execution)
			selector.AddFuture(workflow_execution, func(f workflow.Future) {
				if errCompensation := f.Get(ctx, nil); errCompensation != nil {
					workflow.GetLogger(ctx).Error("Executing compensation failed", errCompensation)
				}
			})
		}
		for range s {
			selector.Select(ctx)
		}

	}
}
