package component

import (
	"context"
	"github.com/go-redis/redis/v8"
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

type RedisV8 struct {
	client *redis.Client
}

type RedisV8Pipeline struct {
	pipeliner redis.Pipeliner
}

type RedisV8Cmd struct {
	args   []interface{}
	result interface{}
	err    error
}

func (r *RedisV8Cmd) Result() (interface{}, error) {
	return r.result, r.err
}

func (r *RedisV8Cmd) Args() []interface{} {
	return r.args
}

func (r *RedisV8Pipeline) Exec(ctx context.Context) ([]Cmder, error) {
	v, e := r.pipeliner.Exec(ctx)
	retCmdList := make([]Cmder, len(v))
	for _, vv := range v {
		cmd := vv.(*redis.Cmd)
		i, e := cmd.Result()
		retCmd := RedisV8Cmd{
			args:   vv.Args(),
			result: i,
			err:    e,
		}
		retCmdList = append(retCmdList, &retCmd)
	}

	return retCmdList, e
}

func (r *RedisV8Pipeline) Eval(ctx context.Context, script string, keys []string, args ...interface{}) Cmder {
	return r.pipeliner.Eval(ctx, script, keys, args...)
}

func (r *RedisV8) Del(ctx context.Context, keys ...string) Cmder {
	v := r.client.Del(ctx, keys...)
	i, e := v.Result()
	retCmd := RedisV8Cmd{
		args:   v.Args(),
		result: i,
		err:    e,
	}
	return &retCmd
}

func (r *RedisV8) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return r.client.Eval(ctx, script, keys, args...).Result()
}

func (r *RedisV8) Pipeline() Pipeliner {
	return &RedisV8Pipeline{
		pipeliner: r.client.Pipeline(),
	}
}
