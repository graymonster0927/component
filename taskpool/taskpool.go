package taskpool

import (
	"context"
	"errors"
	"github.com/panjf2000/ants/v2"
	"sync"
	"time"
)

var gPool *ants.Pool
var gPoolSize = 5000
var once = sync.Once{}

func getGPoolInstance() (*ants.Pool, error) {
	var err error
	once.Do(func() {
		gPool, err = ants.NewPool(gPoolSize, ants.WithNonblocking(true),
			ants.WithExpiryDuration(time.Hour*24))
	})
	return gPool, err
}

type TaskType int

type task struct {
	params   map[string]interface{}
	taskType TaskType
	label    string
}
type taskPool struct {
	taskCount int
	fn        map[TaskType]func(ctx context.Context, params map[string]interface{}) (interface{}, error)
	ctx       context.Context
	taskList  []*task
	errList   map[string]error
	retList   map[string]interface{}
}

func GetTaskPool(ctx context.Context) *taskPool {
	return &taskPool{
		ctx:      ctx,
		taskList: make([]*task, 0, 8),
		errList:  make(map[string]error),
		retList:  make(map[string]interface{}),
		fn:       make(map[TaskType]func(ctx context.Context, params map[string]interface{}) (interface{}, error)),
	}
}

func (t *taskPool) SetGPoolSize(size int) {
	gPoolSize = size
}

func (t *taskPool) SetTaskHandler(taskType TaskType, fn func(ctx context.Context, params map[string]interface{}) (interface{}, error)) {
	t.fn[taskType] = fn
}

func (t *taskPool) AddTask(taskType TaskType, label string, params map[string]interface{}) {
	t.taskList = append(t.taskList, &task{
		taskType: taskType,
		label:    label,
		params:   params,
	})
	t.taskCount++
}

func (t *taskPool) Start() error {

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
		fn, ok := t.fn[task.taskType]
		if !ok {
			errCh <- map[string]error{task.label: errors.New("当前任务类型没有配置处理函数")}
			continue
		}
		wg.Add(1)
		taskCopy := task
		gPool, err := getGPoolInstance()
		if err != nil {
			errCh <- map[string]error{task.label: err}
			wg.Done()
			continue
		}

		if err := gPool.Submit(func() {
			defer func() {
				if r := recover(); r != nil {
					errCh <- map[string]error{taskCopy.label: errors.New("TaskPool Do Task Error")}
				}
				wg.Done()
			}()

			data, err := fn(t.ctx, taskCopy.params)
			errCh <- map[string]error{taskCopy.label: err}
			retCh <- map[string]interface{}{taskCopy.label: data}
		}); err != nil {
			errCh <- map[string]error{task.label: err}
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

func (t *taskPool) GetRetList() map[string]interface{} {
	return t.retList
}

func (t *taskPool) GetErrList() map[string]error {
	return t.errList
}

func (t *taskPool) Clear(ctx context.Context) {
	t.ctx = ctx
	t.taskList = make([]*task, 0, 8)
	t.errList = make(map[string]error)
	t.retList = make(map[string]interface{})
}
