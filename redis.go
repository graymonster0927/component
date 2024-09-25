package component

import (
	"context"
)

type RedisInterface interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
	Pipeline() Pipeliner
	Del(ctx context.Context, keys ...string) Cmder
}

type Pipeliner interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) Cmder
	Exec(ctx context.Context) ([]Cmder, error)
}

type Cmder interface {
	Args() []interface{}
	Result() (interface{}, error)
}

type RedisDefault struct{}

func (r *RedisDefault) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return nil, nil
}

func (r *RedisDefault) Pipeline() Pipeliner {
	return nil
}

func (r *RedisDefault) Del(ctx context.Context, keys ...string) Cmder {
	return nil
}
