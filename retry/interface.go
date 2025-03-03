package retry

import (
	"errors"
	"time"
)

type Strategy int

const (
	RetryMaxTimes Strategy = iota
	RetryUntilSuccess
)

type Interface interface {
	DoRetryWithParams(key string, params ...any) error
	DoRetryReturn(fn func() (interface{}, error)) (interface{}, error)
	DoRetry(fn func() error) error
}

type OptionsInterface interface {
	Apply(ins Interface)
}

type OptionFn func(option OptionsInterface)

var strategyInit = make(map[Strategy]func(options ...OptionFn) (Interface, error))

func GetRetryHelper(strategy Strategy, options ...OptionFn) (Interface, error) {
	if fn, ok := strategyInit[strategy]; ok {
		return fn(options...)
	}

	return nil, errors.New("strategy have not register")
}

func RegisterStrategy(strategy Strategy, fn func(options ...OptionFn) (Interface, error)) error {
	if _, ok := strategyInit[strategy]; ok {
		return errors.New("strategy already registered")
	}

	strategyInit[strategy] = fn
	return nil
}

func init() {

	if err := RegisterStrategy(RetryMaxTimes, func(options ...OptionFn) (Interface, error) {
		ins := NewMaxTimes(3, 100*time.Millisecond, -1, true)

		option := &MaxTimesOption{}
		for _, fn := range options {
			fn(option)
		}
		option.Apply(ins)

		return ins, nil
	}); err != nil {
		panic(err)
	}
}
