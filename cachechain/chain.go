package cachechain

import (
	"context"
	"errors"
	"github.com/graymonster0927/component"
	"github.com/graymonster0927/component/cachechain/cache"
	"github.com/graymonster0927/component/cachechain/cacheerr"
	"github.com/graymonster0927/component/cachechain/helper"
)

type Chain struct {
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

func NewCacheChain() *Chain {
	return &Chain{
		cacheList: make([]cache.CacheInterface, 0),
	}
}

func (c *Chain) WithCache(cache cache.CacheInterface) {
	c.cacheList = append(c.cacheList, cache)
}

func (c *Chain) SetFnGetNoCache(fn func(c context.Context, key string) (string, error)) {
	for _, c := range c.cacheList {
		c.SetFnGetNoCache(fn)
	}
}

func (c *Chain) SetFnBatchGetNoCache(fn func(c context.Context, keyList []string) (map[string]string, error)) {
	for _, c := range c.cacheList {
		c.SetFnBatchGetNoCache(fn)
	}
}

func (c *Chain) SetKeyPrefix(keyPrefix string) {
	for _, c := range c.cacheList {
		c.SetKeyPrefix(keyPrefix)
	}
}

func (c *Chain) Get(ctx context.Context, key string) GetResult {
	ret := GetResult{
		Exist: false,
	}

	if len(c.cacheList) == 0 {
		ret.Err = cacheerr.NoCacheSet
		return ret
	}

	for _, c := range c.cacheList {
		getRet := c.GetFromCache(ctx, key)
		ret.CacheName = c.GetName()
		if getRet.IsSuccess() {
			ret.Err = nil
			ret.FromCache = true
			ret.Exist = true
			ret.V = getRet.Value
			return ret
		} else {
			component.Logger.Errorf(ctx, "cache %s get failed, err: %v", c.GetName(), getRet.Err)
			ret.Err = errors.Join(ret.Err, getRet.Err)
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

	return ret
}

func (c *Chain) BatchGet(ctx context.Context, keyList []string) map[string]GetResult {
	ret := make(map[string]GetResult)

	if len(c.cacheList) == 0 {
		for _, key := range keyList {
			ret[key] = GetResult{
				ErrHelper: helper.ErrHelper{Err: cacheerr.NoCacheSet},
			}
		}
		return ret
	}

	for _, c := range c.cacheList {
		getRetMap := c.BatchGetFromCache(ctx, keyList)
		keyList = make([]string, 0, len(keyList))
		for key, getRet := range getRetMap {
			if getRet.IsSuccess() {
				ret[key] = GetResult{
					CacheName: c.GetName(),
					ErrHelper: helper.ErrHelper{Err: nil},
					FromCache: true,
					Exist:     getRet.Exist,
					V:         getRet.Value,
				}
				continue
			} else {
				component.Logger.Errorf(ctx, "cache %s get failed, err: %v", c.GetName(), getRet.Err)
				var preErr error
				if _, ok := ret[key]; ok {
					preErr = ret[key].Err
				}
				ret[key] = GetResult{
					CacheName: c.GetName(),
					ErrHelper: helper.ErrHelper{Err: errors.Join(preErr, getRet.Err)},
					FromCache: false,
					Exist:     false,
					V:         "",
				}
			}

			switch getRet.HandleErrStrategy {
			case cache.HandleErrStrategyContinue:
				keyList = append(keyList, key)
			case cache.HandleErrStrategyBreak:
				ret[key] = GetResult{
					CacheName: c.GetName(),
					ErrHelper: helper.ErrHelper{Err: getRet.Err},
					FromCache: false,
					Exist:     false,
					V:         "",
				}
			case cache.HandleErrStrategyRetry:
				getRet = c.RetryGetFromCache(ctx, key)
				if getRet.IsSuccess() {
					ret[key] = GetResult{
						CacheName: c.GetName(),
						ErrHelper: helper.ErrHelper{Err: nil},
						FromCache: true,
						Exist:     getRet.Exist,
						V:         getRet.Value,
					}
				}
			}
		}
	}
	return ret
}

func (c *Chain) Set(ctx context.Context, key string, val string) SetResult {
	ret := SetResult{}

	if len(c.cacheList) == 0 {
		ret.Err = cacheerr.NoCacheSet
		return ret
	}

	for _, c := range c.cacheList {
		setRet := c.SetCache(ctx, key, val)
		if setRet.IsSuccess() {
			continue
		} else {
			component.Logger.Errorf(ctx, "cache %s set failed, err: %v", c.GetName(), setRet.Err)
		}

		switch setRet.HandleErrStrategy {
		case cache.HandleErrStrategyContinue:
			continue
		case cache.HandleErrStrategyBreak:
			ret.Err = setRet.Err
			return ret
		case cache.HandleErrStrategyRetry:
			setRet = c.RetrySetCache(ctx, key, val)
			if setRet.IsSuccess() {
				continue
			} else {
				ret.Err = setRet.Err
				return ret
			}
		}
	}

	return ret

}

func (c *Chain) BatchSet(ctx context.Context, keyList []string, valList []string) map[string]SetResult {
	ret := make(map[string]SetResult)

	if len(c.cacheList) == 0 {
		for _, key := range keyList {
			ret[key] = SetResult{
				ErrHelper: helper.ErrHelper{Err: cacheerr.NoCacheSet},
			}
		}
		return ret
	}

	//todo 批量数校验
	vMap := make(map[string]string)
	for i, key := range keyList {
		vMap[key] = valList[i]
	}

	for _, c := range c.cacheList {
		setRetMap := c.BatchSetCache(ctx, keyList, valList)
		keyList = make([]string, 0, len(keyList))
		for key, setRet := range setRetMap {
			if setRet.IsSuccess() {
				ret[key] = SetResult{
					ErrHelper: helper.ErrHelper{Err: nil},
				}
				keyList = append(keyList, key)
				continue
			} else {
				component.Logger.Errorf(ctx, "cache %s set failed, err: %v", c.GetName(), setRet.Err)
			}

			switch setRet.HandleErrStrategy {
			case cache.HandleErrStrategyContinue:
				ret[key] = SetResult{
					ErrHelper: helper.ErrHelper{Err: nil},
				}
				keyList = append(keyList, key)
			case cache.HandleErrStrategyBreak:
				ret[key] = SetResult{
					ErrHelper: helper.ErrHelper{Err: setRet.Err},
				}
			case cache.HandleErrStrategyRetry:
				setRet = c.RetrySetCache(ctx, key, vMap[key])
				if setRet.IsSuccess() {
					ret[key] = SetResult{
						ErrHelper: helper.ErrHelper{Err: nil},
					}
					keyList = append(keyList, key)
				} else {
					ret[key] = SetResult{
						ErrHelper: helper.ErrHelper{Err: setRet.Err},
					}
				}
			}
		}
	}

	return ret
}

func (c *Chain) Clear(ctx context.Context, key string) ClearResult {
	ret := ClearResult{}

	if len(c.cacheList) == 0 {
		ret.Err = cacheerr.NoCacheSet
		return ret
	}

	for _, c := range c.cacheList {
		clearRet := c.ClearCache(ctx, key)
		if clearRet.IsSuccess() {
			continue
		} else {
			component.Logger.Errorf(ctx, "cache %s clear failed, err: %v", c.GetName(), clearRet.Err)
		}

		switch clearRet.HandleErrStrategy {
		case cache.HandleErrStrategyContinue:
			continue
		case cache.HandleErrStrategyBreak:
			ret.Err = clearRet.Err
			return ret
		case cache.HandleErrStrategyRetry:
			clearRet = c.RetryClearCache(ctx, key)
			if clearRet.IsSuccess() {
				continue
			} else {
				ret.Err = clearRet.Err
				return ret
			}
		}
	}

	return ret
}

func (c *Chain) BatchClear(ctx context.Context, keyList []string) map[string]ClearResult {
	ret := make(map[string]ClearResult)

	if len(c.cacheList) == 0 {
		for _, key := range keyList {
			ret[key] = ClearResult{
				ErrHelper: helper.ErrHelper{Err: cacheerr.NoCacheSet},
			}
		}
		return ret
	}

	for _, c := range c.cacheList {
		clearRetMap := c.BatchClearCache(ctx, keyList)
		keyList = make([]string, 0, len(keyList))
		for key, clearRet := range clearRetMap {
			if clearRet.IsSuccess() {
				ret[key] = ClearResult{
					ErrHelper: helper.ErrHelper{Err: nil},
				}
				keyList = append(keyList, key)
				continue
			} else {
				component.Logger.Errorf(ctx, "cache %s clear failed, err: %v", c.GetName(), clearRet.Err)
			}

			switch clearRet.HandleErrStrategy {
			case cache.HandleErrStrategyContinue:
				ret[key] = ClearResult{
					ErrHelper: helper.ErrHelper{Err: nil},
				}
				keyList = append(keyList, key)
			case cache.HandleErrStrategyBreak:
				ret[key] = ClearResult{
					ErrHelper: helper.ErrHelper{Err: clearRet.Err},
				}
			case cache.HandleErrStrategyRetry:
				clearRet = c.RetryClearCache(ctx, key)
				if clearRet.IsSuccess() {
					ret[key] = ClearResult{
						ErrHelper: helper.ErrHelper{Err: nil},
					}
					keyList = append(keyList, key)
				} else {
					ret[key] = ClearResult{
						ErrHelper: helper.ErrHelper{Err: clearRet.Err},
					}
				}
			}
		}
	}

	return ret
}
