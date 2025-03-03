package retry

import (
	"errors"
	"golang.org/x/exp/rand"
	"math"
	"time"
)

type MaxTimes struct {
	MaxTimes      int
	RetryTimeout  time.Duration
	MaxDelay      time.Duration
	IsExponential bool
}

type MaxTimesOption struct {
	MaxTimes      int
	RetryTimeout  time.Duration
	MaxDelay      time.Duration
	IsExponential bool
}

func (m *MaxTimesOption) Apply(ins Interface) {
	if v, ok := ins.(*MaxTimes); ok {
		if m.MaxTimes > 0 {
			v.MaxTimes = m.MaxTimes
		}

		if m.RetryTimeout > 0 {
			v.RetryTimeout = m.RetryTimeout
		}

		if m.MaxDelay > 0 {
			v.MaxDelay = m.MaxDelay
		}

		if !m.IsExponential {
			v.IsExponential = m.IsExponential
		}
	}
}
func WithMaxTimesMaxTimes(maxTimes int) OptionFn {
	return func(option OptionsInterface) {
		if v, ok := option.(*MaxTimesOption); ok {
			v.MaxTimes = maxTimes
		}
	}
}

func WithMaxTimesRetryTimeout(retryTimeout time.Duration) OptionFn {
	return func(option OptionsInterface) {
		if v, ok := option.(*MaxTimesOption); ok {
			v.RetryTimeout = retryTimeout
		}
	}
}

func WithMaxTimesMaxDelay(maxDelay time.Duration) OptionFn {
	return func(option OptionsInterface) {
		if v, ok := option.(*MaxTimesOption); ok {
			v.MaxDelay = maxDelay
		}
	}
}

func WithMaxTimesIsExponential(isExponential bool) OptionFn {
	return func(option OptionsInterface) {
		if v, ok := option.(*MaxTimesOption); ok {
			v.IsExponential = isExponential
		}
	}
}

func NewMaxTimes(maxTimes int, retryTimeout, maxDelay time.Duration, isExponential bool) *MaxTimes {
	return &MaxTimes{
		MaxTimes:      maxTimes,
		RetryTimeout:  retryTimeout,
		MaxDelay:      maxDelay,
		IsExponential: isExponential,
	}
}

func (m *MaxTimes) DoRetry(fn func() error) error {
	var err error
	retryTimes := 1
	for i := 0; i < m.MaxTimes; i++ {
		time.Sleep(m.getRetryTimeout(retryTimes))
		if err = fn(); err != nil {
			retryTimes++
			continue
		}
		return nil
	}
	return err
}

func (u *MaxTimes) DoRetryWithParams(key string, params ...any) error {
	return errors.New("this kind strategy not support retry with params")
}

func (m *MaxTimes) DoRetryReturn(fn func() (interface{}, error)) (interface{}, error) {
	var err error
	var ret interface{}
	retryTimes := 1
	for i := 0; i < m.MaxTimes; i++ {
		time.Sleep(m.getRetryTimeout(retryTimes))
		if ret, err = fn(); err != nil {
			retryTimes++
			continue
		} else {
			return ret, nil
		}
	}
	return ret, err
}

func (m *MaxTimes) getRetryTimeout(retryTimes int) time.Duration {
	if m.IsExponential {
		timeout := m.RetryTimeout * time.Duration(math.Pow(2, float64(retryTimes-1)))
		// 添加随机抖动（jitter）
		jitter := time.Duration(rand.Int63n(int64(timeout / 2)))
		timeout = timeout/2 + jitter
		if m.MaxDelay > 0 && timeout > m.MaxDelay {
			timeout = m.MaxDelay
		}

		return timeout
	}

	return m.RetryTimeout
}
