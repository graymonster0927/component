package taskpool

import (
	"context"
	"errors"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"sync"
	"time"
)

var GPool *ants.Pool
var GPoolSize = 5000

var once = sync.Once{}

func GetGPoolInstance() *ants.Pool {
	once.Do(func() {
		GPool, _ = ants.NewPool(GPoolSize, ants.WithNonblocking(true),
			ants.WithExpiryDuration(time.Hour*24))
	})
	return GPool
}

type TaskType int

type Task struct {
	Params   map[string]interface{}
	TaskType TaskType
	Label    string
}

type TaskPool struct {
	taskCount int
	fn        map[TaskType]func(ctx *context.Context, params map[string]interface{}) (interface{}, error)
	ctx       *context.Context
	taskList  []*Task
	errList   map[string]error
	retList   map[string]interface{}
}

func GetTaskPool(ctx *context.Context) *TaskPool {
	return &TaskPool{
		ctx:      ctx,
		taskList: make([]*Task, 0, 8),
		errList:  make(map[string]error),
		retList:  make(map[string]interface{}),
		fn:       make(map[TaskType]func(ctx *context.Context, params map[string]interface{}) (interface{}, error)),
	}
}

func (t *TaskPool) SetTaskHandler(taskType TaskType, fn func(ctx *context.Context, params map[string]interface{}) (interface{}, error)) {
	t.fn[taskType] = fn
}

func (t *TaskPool) AddTask(taskType TaskType, label string, params map[string]interface{}) {
	t.taskList = append(t.taskList, &Task{
		TaskType: taskType,
		Label:    label,
		Params:   params,
	})
	t.taskCount++
}

func (t *TaskPool) Start() error {

	if t.taskCount == 0 {
		return errors.New("当前没有任务")
	}

	if t.fn == nil {
		return errors.New("当前没有配置处理任务函数")
	}

	//todo  ratelimit
	wg := sync.WaitGroup{}
	errCh := make(chan map[string]error, t.taskCount*2)
	retCh := make(chan map[string]interface{}, t.taskCount)

	for _, task := range t.taskList {
		fn, ok := t.fn[task.TaskType]
		if !ok {
			errCh <- map[string]error{task.Label: errors.New("当前任务类型没有配置处理函数")}
			continue
		}
		wg.Add(1)
		taskCopy := task
		if err := GetGPoolInstance().Submit(func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println(t.ctx, "TaskPool Do Task Error", r)
					errCh <- map[string]error{taskCopy.Label: errors.New("TaskPool Do Task Error")}
				}
				wg.Done()
			}()

			data, err := fn(t.ctx, taskCopy.Params)
			errCh <- map[string]error{taskCopy.Label: err}
			retCh <- map[string]interface{}{taskCopy.Label: data}
		}); err != nil {
			errCh <- map[string]error{task.Label: err}
			wg.Done()
		}
	}

	wg.Wait()

	for len(errCh) > 0 {
		tmp := <-errCh
		for idx, item := range tmp {
			t.errList[idx] = item
		}
	}
	close(errCh)

	for len(retCh) > 0 {
		tmp := <-retCh
		for idx, item := range tmp {
			t.retList[idx] = item
		}
	}
	close(retCh)

	return nil
}

func (t *TaskPool) GetRetList() map[string]interface{} {
	return t.retList
}

func (t *TaskPool) GetErrList() map[string]error {
	return t.errList
}

func (t *TaskPool) Clear(ctx *context.Context) {
	t.ctx = ctx
	t.taskList = make([]*Task, 0, 8)
	t.errList = make(map[string]error)
	t.retList = make(map[string]interface{})
}
