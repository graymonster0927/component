package cache

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/graymonster0927/component"
	"github.com/graymonster0927/component/cachechain/helper"
	"go.uber.org/zap"
)


type MemoryCacheOption func(*memoryOptions)

type memoryOptions struct {
	expireTime     int
	tokenPrefix    string
	strategy       HandleErrStrategy
	conn           component.MemoryInterface
}

func WithMemoryExpireTime(expireTime int) MemoryCacheOption {
	return func(o *memoryOptions) {
		o.expireTime = expireTime
	}
}


func WithMemoryTokenPrefix(tokenPrefix string) MemoryCacheOption {
	return func(o *memoryOptions) {
		o.tokenPrefix = tokenPrefix
	}
}

func WithMemoryHandleErrStrategy(strategy HandleErrStrategy) MemoryCacheOption {
	return func(o *memoryOptions) {
		o.strategy = strategy
	}
}

func WithMemoryConn(conn component.MemoryInterface) MemoryCacheOption {
	return func(o *memoryOptions) {
		o.conn = conn
	}
}

type MemoryCache struct {
	opts      memoryOptions
	fn        func(c context.Context, key string) (string, error)
	batchFn   func(c context.Context, keyList []string) (map[string]string, error)
	keyPrefix string
	conn      component.MemoryInterface
}

func NewMemoryCache(opts ...MemoryCacheOption) *MemoryCache {
	// 创建一个默认的options
	op := memoryOptions{
		expireTime:     7 * 86400,
		tokenPrefix:    "graymonster-cachechain-memory-token",
		strategy:       HandleErrStrategyRetry,
	}
	// 调用动态传入的参数进行设置值
	for _, option := range opts {
		option(&op)
	}
	cache := &MemoryCache{
		opts: op,
		//todo
		//conn: &component.RedisDefault{},
	}
	if op.conn != nil {
		cache.conn = op.conn
	}
	return cache
}

func (r *MemoryCache) GetName() string {
	return reflect.TypeOf(r).String()
}

func (r *MemoryCache) SetFnGetNoCache(fn func(c context.Context, key string) (string, error)) {
	r.fn = fn
}

func (r *MemoryCache) SetFnBatchGetNoCache(fn func(c context.Context, keyList []string) (map[string]string, error)) {
	r.batchFn = fn
}

func (r *MemoryCache) SetKeyPrefix(keyPrefix string) {
	r.keyPrefix = keyPrefix
}

func (r *MemoryCache) GetFromCache(c context.Context, key string) GetCacheResult {
	prefixKey := fmt.Sprintf(r.keyPrefix, key)
	val, err := r.conn.Get(c, prefixKey)
	if err != nil {
		component.Logger.Errorf(c, "memory get key failed", zap.Error(err), zap.String("key", prefixKey))
	}
	return GetCacheResult{
		Value:             val,
		Exist:             val != "",
		ErrHelper:         helper.ErrHelper{
			Err: err,
		},
		HandleErrStrategy: r.opts.strategy,
	}
}

func (r *MemoryCache) BatchGetFromCache(c context.Context, keyList []string) map[string]GetCacheResult {
	retMap := make(map[string]GetCacheResult)
	for _, key := range keyList {
		retMap[key] = r.GetFromCache(c, key)
	}
	return retMap
}

func (r *MemoryCache) SetCache(ctx context.Context, key string, val string) SetCacheResult {
	prefixKey := fmt.Sprintf(r.keyPrefix, key)
	err := r.conn.Set(ctx, prefixKey, val, time.Duration(r.opts.expireTime)*time.Second)
	if err != nil {
		component.Logger.Errorf(ctx, "memory set key failed", zap.Error(err), zap.String("key", prefixKey))
	}
	return SetCacheResult{
		ErrHelper: helper.ErrHelper{
			Err: err,
		},
		HandleErrStrategy: r.opts.strategy,
	}
}

func (r *MemoryCache) BatchSetCache(ctx context.Context, keyList []string, valList []string) map[string]SetCacheResult {
	retMap := make(map[string]SetCacheResult)
	for idx, key := range keyList {
		ret := r.SetCache(ctx, key, valList[idx])
		retMap[keyList[idx]] = ret
	}
	return retMap
}

func (r *MemoryCache) ClearCache(ctx context.Context, key string) ClearCacheResult {
	prefixKey := fmt.Sprintf(r.keyPrefix, key)
	err := r.conn.Del(ctx, prefixKey)
	if err != nil {
		component.Logger.Errorf(ctx, "redis clear key failed", zap.Error(err), zap.String("key", prefixKey))
	}
	return ClearCacheResult{
		ErrHelper: helper.ErrHelper{
			Err: err,
		},
		HandleErrStrategy: r.opts.strategy,
	}
}

func (r *MemoryCache) BatchClearCache(ctx context.Context, keyList []string) map[string]ClearCacheResult {
	retMap := make(map[string]ClearCacheResult)
	for i := 0; i < len(keyList); i++ {
		ret := r.ClearCache(ctx, keyList[i])
		retMap[keyList[i]] = ret
	}
	return retMap
}

func (r *MemoryCache) RetryGetFromCache(ctx context.Context, key string) GetCacheResult {
	//todo implement
	return GetCacheResult{
		HandleErrStrategy: HandleErrStrategyBreak,
	}
}

func (r *MemoryCache) RetrySetCache(ctx context.Context, key string, val string) SetCacheResult {
	//todo implement
	return SetCacheResult{
		HandleErrStrategy: HandleErrStrategyBreak,
	}
}

func (r *MemoryCache) RetryClearCache(ctx context.Context, key string) ClearCacheResult {
	//todo implement
	return ClearCacheResult{
		HandleErrStrategy: HandleErrStrategyBreak,
	}
}

