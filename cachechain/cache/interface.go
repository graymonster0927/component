package cache

import (
	"context"
	"github.com/graymonster0927/component/cachechain/helper"
)

//值
//是否存在
//是否成功
//失败策略

type HandleErrStrategy int
type CacheType int

const (
	_ = iota
	// HandleErrStrategyContinue 继续
	HandleErrStrategyContinue HandleErrStrategy = iota

	// HandleErrStrategyRollback 回滚
	HandleErrStrategyRollback HandleErrStrategy = iota

	// HandleErrStrategyRetry 重试
	HandleErrStrategyRetry HandleErrStrategy = iota

	// HandleErrStrategyBreak 中断
	HandleErrStrategyBreak HandleErrStrategy = iota
)

type GetCacheResult struct {
	Value string
	Exist bool
	helper.ErrHelper
	HandleErrStrategy HandleErrStrategy
}

type SetCacheResult struct {
	helper.ErrHelper
	HandleErrStrategy HandleErrStrategy
}

type ClearCacheResult struct {
	helper.ErrHelper
	HandleErrStrategy HandleErrStrategy
}

type CacheInterface interface {
	GetFromCache(ctx context.Context, key string) GetCacheResult
	BatchGetFromCache(c context.Context, keyList []string) map[string]GetCacheResult
	SetCache(ctx context.Context, key string, val string) SetCacheResult
	BatchSetCache(ctx context.Context, keyList []string, valList []string) map[string]SetCacheResult
	ClearCache(ctx context.Context, key string) ClearCacheResult
	BatchClearCache(ctx context.Context, keyList []string) map[string]ClearCacheResult
	RetryGetFromCache(ctx context.Context, key string) GetCacheResult
	RetrySetCache(ctx context.Context, key string, val string) SetCacheResult
	RetryClearCache(ctx context.Context, key string) ClearCacheResult
	GetName() string
	SetFnGetNoCache(fn func(c context.Context, key string) (string, error))
	SetFnBatchGetNoCache(fn func(c context.Context, keyList []string) (map[string]string, error))
	SetKeyPrefix(keyPrefix string)
}
