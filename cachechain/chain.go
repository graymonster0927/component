package cachechain

import (
	"context"
	"github.com/graymonster0927/component/cachechain/cache"
	"github.com/graymonster0927/component/cachechain/cacheerr"
	"github.com/graymonster0927/component/cachechain/helper"
)

type chain struct {
	cacheList []cache.CacheInterface
}

type GetResult struct {
	helper.ErrHelper
	V         string
	Exist     bool
	FromCache bool
	CacheName string
}

type SetResult struct {
	helper.ErrHelper
}

type ClearResult struct {
	helper.ErrHelper
}

func NewCacheChain() *chain {
	return &chain{
		cacheList: make([]cache.CacheInterface, 0),
	}
}

func (c *chain) WithCache(cache cache.CacheInterface) {
	c.cacheList = append(c.cacheList, cache)
}

func (c *chain) Get(ctx context.Context, key string, fnGetNoCache func(key string) (string, error)) GetResult {
	ret := GetResult{}

	if len(c.cacheList) == 0 {
		ret.Err = cacheerr.NoCacheSet
		return ret
	}

	for _, c := range c.cacheList {
		getRet := c.GetFromCache(ctx, key)
		if getRet.IsSuccess() {
			ret.CacheName = c.GetName()
			ret.FromCache = true
			ret.Exist = getRet.Exist
			ret.V = getRet.Value
			ret.Err = nil
			return ret
		}

		switch getRet.HandleErrStrategy {
		case cache.HandleErrStrategyContinue:
			continue
		case cache.HandleErrStrategyBreak:
			ret.Err = getRet.Err
			return ret
		case cache.HandleErrStrategyRetry:
			getRet = c.RetryGetFromCache(ctx, key)
			if getRet.IsSuccess() {
				ret.CacheName = c.GetName()
				ret.FromCache = true
				ret.Exist = getRet.Exist
				ret.V = getRet.Value
				ret.Err = nil
				return ret
			}
		}
	}

	//到这里都没有拿到
	v, err := fnGetNoCache(key)
	ret.Err = err
	ret.V = v
	ret.Exist = ret.IsSuccess()
	ret.FromCache = false
	return ret
}

//func (c *chain) BatchGet(ctx context.Context, key string, fnGetNoCache func(keyList []string) (map[string]string, error)) map[string]GetResult {
//
//}
//
//func (c *chain) Set(ctx context.Context, key string, val string) SetResult {
//
//}
//
//func (c *chain) BatchSet(ctx context.Context, keyList []string, valList []string) map[string]SetResult {
//
//}
//
//func (c *chain) Clear(ctx context.Context, key string) ClearResult {
//
//}
//
//func (c *chain) BatchClear(ctx context.Context, keyList []string) map[string]ClearResult {
//
//}
