package taskpool

import (
	"context"
)

type TaskHandlerI interface {
	GetTaskFn() func(ctx context.Context, params map[string]interface{}) (interface{}, error)
	GetTaskType() TaskType
}
