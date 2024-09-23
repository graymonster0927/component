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

type getCacheResult struct {
	Value string
	Exist bool
	helper.ErrHelper
	HandleErrStrategy HandleErrStrategy
}

type setCacheResult struct {
	helper.ErrHelper
	HandleErrStrategy HandleErrStrategy
}

type clearCacheResult struct {
	helper.ErrHelper
	HandleErrStrategy HandleErrStrategy
}

type CacheInterface interface {
	GetFromCache(ctx context.Context, key string) getCacheResult
	BatchGetFromCache(c context.Context, keyList []string) map[string]getCacheResult
	SetCache(ctx context.Context, key string, val string) setCacheResult
	BatchSetCache(ctx context.Context, keyList []string, valList []string) map[string]setCacheResult
	ClearCache(ctx context.Context, key string) clearCacheResult
	BatchClearCache(ctx context.Context, keyList []string) map[string]clearCacheResult
	RetryGetFromCache(ctx context.Context, key string) getCacheResult
	RetryBatchGetFromCache(ctx context.Context, keyList []string) map[string]getCacheResult
	GetName() string
}

//type CommonCacheOption struct {
//	CacheExpire int
//}
//type option func(*CommonCacheOption)
//
//func (c *CommonCacheOption) WithExpireTime(expireTime int) option {
//	return func(o *CommonCacheOption) {
//		o.CacheExpire = expireTime
//	}
//}
