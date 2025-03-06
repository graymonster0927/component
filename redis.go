package component

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisInterface interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
	Pipeline() Pipeliner
	Del(ctx context.Context, keys ...string) Cmder
	Get(ctx context.Context, key string) (string, error)
	//LPush(ctx context.Context, key string, values ...interface{}) *IntCmd
	//RPop(ctx context.Context, key string) *StringCmd
}

type Pipeliner interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) Cmder
	Exec(ctx context.Context) ([]Cmder, error)
}

type Cmder interface {
	Args() []interface{}
	Result() (interface{}, error)
}

type baseCmd struct {
	ctx    context.Context
	args   []interface{}
	err    error
	keyPos int8

	_readTimeout *time.Duration
}

type StringCmd struct {
	baseCmd

	val string
}

func (cmd *StringCmd) Result() (string, error) {
	return cmd.val, cmd.err
}

type IntCmd struct {
	baseCmd

	val int64
}

func (cmd *IntCmd) Result() (int64, error) {
	return cmd.val, cmd.err
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

func (r *RedisDefault) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (r *RedisDefault) LPush(ctx context.Context, key string, values ...interface{}) *IntCmd {
	return nil
}

func (r *RedisDefault) RPop(ctx context.Context, key string) *StringCmd {
	return nil
}

type RedisV8 struct {
	Client *redis.Client
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
	retCmdList := make([]Cmder, 0, len(v))
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
	v := r.Client.Del(ctx, keys...)
	i, e := v.Result()
	retCmd := RedisV8Cmd{
		args:   v.Args(),
		result: i,
		err:    e,
	}
	return &retCmd
}

func (r *RedisV8) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return r.Client.Eval(ctx, script, keys, args...).Result()
}

func (r *RedisV8) Pipeline() Pipeliner {
	return &RedisV8Pipeline{
		pipeliner: r.Client.Pipeline(),
	}
}

func (r *RedisV8) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

func (r *RedisV8) LPush(ctx context.Context, key string, values ...interface{}) *IntCmd {
	retCmd := r.Client.LPush(ctx, key, values...)
	return &IntCmd{
		baseCmd: baseCmd{
			args: retCmd.Args(),
		},
		val: retCmd.Val(),
	}
}

func (r *RedisV8) RPop(ctx context.Context, key string) *StringCmd {
	retCmd := r.Client.RPop(ctx, key)
	return &StringCmd{
		baseCmd: baseCmd{
			args: retCmd.Args(),
		},
		val: retCmd.Val(),
	}
}
