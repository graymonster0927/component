package retry

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/graymonster0927/component"
	"go.uber.org/zap"
	"math"
	"math/rand"
	"time"
)

type UntilSuccess struct {
	logger        component.LoggerInterface
	redis         component.RedisInterface
	MaxTimes      int
	RetryTimeout  time.Duration
	MaxDelay      time.Duration
	IsExponential bool
	retryMethod   map[string]func(params ...any) error
}

func NewUntilSuccess(maxTimes int, retryTimeout, maxDelay time.Duration, isExponential bool) *UntilSuccess {
	return &UntilSuccess{
		MaxTimes:      maxTimes,
		RetryTimeout:  retryTimeout,
		MaxDelay:      maxDelay,
		IsExponential: isExponential,
		retryMethod:   make(map[string]func(...any) error),
	}
}

type UntilSuccessOption struct {
	logger        component.LoggerInterface
	redis         component.RedisInterface
	MaxTimes      int
	RetryTimeout  time.Duration
	MaxDelay      time.Duration
	IsExponential bool
}

type RetryTask struct {
	Key        string
	Params     []any
	RetryTimes int
	StartRetry int64
}

func (u *UntilSuccessOption) Apply(ins Interface) {
	if v, ok := ins.(*UntilSuccess); ok {
		if u.logger != nil {
			v.logger = u.logger
		}

		if u.redis != nil {
			v.redis = u.redis
		}

		if u.MaxTimes > 0 {
			v.MaxTimes = u.MaxTimes
		}

		if u.RetryTimeout > 0 {
			v.RetryTimeout = u.RetryTimeout
		}

		if u.MaxDelay > 0 {
			v.MaxDelay = u.MaxDelay
		}

		if !u.IsExponential {
			v.IsExponential = u.IsExponential
		}
	}
}
func WithUntilSuccessLogger(logger component.LoggerInterface) OptionFn {
	return func(option OptionsInterface) {
		if v, ok := option.(*UntilSuccessOption); ok {
			v.logger = logger
		}
	}
}
func WithUntilSuccessRedis(redis component.RedisInterface) OptionFn {
	return func(option OptionsInterface) {
		if v, ok := option.(*UntilSuccessOption); ok {
			v.redis = redis
		}
	}
}

func WithUntilSuccessMaxTimes(maxTimes int) OptionFn {
	return func(option OptionsInterface) {
		if v, ok := option.(*UntilSuccessOption); ok {
			v.MaxTimes = maxTimes
		}
	}
}

func WithUntilSuccessRetryTimeout(retryTimeout time.Duration) OptionFn {
	return func(option OptionsInterface) {
		if v, ok := option.(*UntilSuccessOption); ok {
			v.RetryTimeout = retryTimeout
		}
	}
}

func WithUntilSuccessMaxDelay(maxDelay time.Duration) OptionFn {
	return func(option OptionsInterface) {
		if v, ok := option.(*UntilSuccessOption); ok {
			v.MaxDelay = maxDelay
		}
	}
}

func WithUntilSuccessIsExponential(isExponential bool) OptionFn {
	return func(option OptionsInterface) {
		if v, ok := option.(*UntilSuccessOption); ok {
			v.IsExponential = isExponential
		}
	}
}

func SetUntilSuccessRetryMethod(ins Interface, key string, fn func(params ...any) error) {
	if v, ok := ins.(*UntilSuccess); ok {
		v.retryMethod[key] = fn
	}
}

func (u *UntilSuccess) DoRetryWithParams(key string, params ...any) error {
	//存储
	task := RetryTask{
		Key:        key,
		Params:     params,
		RetryTimes: 0,
	}
	u.redis.LPush(context.Background(), "retry_task", task)
	return nil
}
func (u *UntilSuccess) DoRetry(fn func() error) error {
	return errors.New("this kind strategy not support retry without params")
}

func (u *UntilSuccess) DoRetryReturn(fn func() (interface{}, error)) (interface{}, error) {
	return nil, errors.New("this kind strategy not support retry without params")
}

func (u *UntilSuccess) getRetryTimeout(retryTimes int) time.Duration {
	if u.IsExponential {
		timeout := u.RetryTimeout * time.Duration(math.Pow(2, float64(retryTimes-1)))
		// 添加随机抖动（jitter）
		jitter := time.Duration(rand.Int63n(int64(timeout / 2)))
		timeout = timeout/2 + jitter
		if u.MaxDelay > 0 && timeout > u.MaxDelay {
			timeout = u.MaxDelay
		}

		return timeout
	}

	return u.RetryTimeout
}

func (u *UntilSuccess) Run() {
	for {

		taskI, err := u.redis.RPop(context.Background(), "retry_task").Result()
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		if taskI == "" {
			time.Sleep(3 * time.Second)
			continue
		}

		var task RetryTask
		if err := json.Unmarshal([]byte(taskI), &task); err != nil {
			u.logger.Errorf(context.Background(), "redis RPop invalid", zap.Error(err), zap.Any("task", task))
			time.Sleep(1 * time.Second)
		}

		if task.RetryTimes >= u.MaxTimes {
			continue
		}

		if task.StartRetry < time.Now().Unix() {
			u.redis.LPush(context.Background(), "retry_task", task)
			time.Sleep(50 * time.Millisecond)
		}

		fn := u.retryMethod[task.Key]
		if err := fn(task.Params...); err != nil {
			u.logger.Errorf(context.Background(), "retry method error", zap.Error(err), zap.Any("task", task))
			task.RetryTimes++
			task.StartRetry = time.Now().Unix() + int64(u.getRetryTimeout(task.RetryTimes).Seconds())
			u.redis.LPush(context.Background(), "retry_task", task)

		}
	}
}
