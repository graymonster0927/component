package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/graymonster0927/component"
	"github.com/graymonster0927/component/cachechain/helper"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	RedisCacheStatusDoReadDB = 1
	RedisCacheStatusWait     = 2
	RedisCacheStatusOK       = 3
)

type redisGetResult struct {
	helper.ErrHelper
	Exist bool
	Value string
	Token string
	//1 说明当前拿到了token
	//2 说明当前已经有请求在走DB
	//3 说明有缓存值
	Status uint8
}

type RedisCacheOption func(*options)

type options struct {
	expireTime     int
	maxWaitingLoop int
	tokenPrefix    string
	strategy       HandleErrStrategy
	conn           component.RedisInterface
}

func WithExpireTime(expireTime int) RedisCacheOption {
	return func(o *options) {
		o.expireTime = expireTime
	}
}

func WithMaxWaitingLoop(maxWaitingLoop int) RedisCacheOption {
	return func(o *options) {
		o.maxWaitingLoop = maxWaitingLoop
	}
}

func WithTokenPrefix(tokenPrefix string) RedisCacheOption {
	return func(o *options) {
		o.tokenPrefix = tokenPrefix
	}
}

func WithHandleErrStrategy(strategy HandleErrStrategy) RedisCacheOption {
	return func(o *options) {
		o.strategy = strategy
	}
}

func WithRedisConn(conn component.RedisInterface) RedisCacheOption {
	return func(o *options) {
		o.conn = conn
	}
}

type RedisCache struct {
	opts      options
	fn        func(c context.Context, key string) (string, error)
	batchFn   func(c context.Context, keyList []string) (map[string]string, error)
	keyPrefix string
	conn      component.RedisInterface
}

func NewRedisCache(opts ...RedisCacheOption) *RedisCache {
	// 创建一个默认的options
	op := options{
		expireTime:     7 * 86400,
		maxWaitingLoop: 5,
		tokenPrefix:    "graymonster-cachechain-redis-token",
		strategy:       HandleErrStrategyContinue,
	}
	// 调用动态传入的参数进行设置值
	for _, option := range opts {
		option(&op)
	}
	cache := &RedisCache{
		opts: op,
		conn: &component.RedisDefault{},
	}
	if op.conn != nil {
		cache.conn = op.conn
	}
	return cache
}

func (r *RedisCache) GetName() string {
	return reflect.TypeOf(r).String()
}

func (r *RedisCache) SetFnGetNoCache(fn func(c context.Context, key string) (string, error)) {
	r.fn = fn
}

func (r *RedisCache) SetFnBatchGetNoCache(fn func(c context.Context, keyList []string) (map[string]string, error)) {
	r.batchFn = fn
}

func (r *RedisCache) SetKeyPrefix(keyPrefix string) {
	r.keyPrefix = keyPrefix
}

func (r *RedisCache) GetFromCache(c context.Context, key string) GetCacheResult {
	getRet := r.getFromRedis(c, key)
	retMap := r.handleGetFromRedis(c, false, map[string]redisGetResult{
		key: getRet,
	}, 0)
	return retMap[key]
}

func (r *RedisCache) BatchGetFromCache(c context.Context, keyList []string) map[string]GetCacheResult {
	getRet := r.batchGetFromRedis(c, keyList)
	return r.handleGetFromRedis(c, true, getRet, 0)
}

func (r *RedisCache) SetCache(ctx context.Context, key string, val string) SetCacheResult {
	prefixKey := fmt.Sprintf(r.keyPrefix, key)
	_, err := r.conn.Del(ctx, prefixKey).Result()
	if err != nil {
		component.Logger.Errorf(ctx, "redis set key failed", zap.Error(err), zap.String("key", prefixKey))
	}
	return SetCacheResult{
		ErrHelper: helper.ErrHelper{
			Err: err,
		},
		HandleErrStrategy: r.opts.strategy,
	}
}

func (r *RedisCache) BatchSetCache(ctx context.Context, keyList []string, valList []string) map[string]SetCacheResult {
	prefixKeyList := make([]string, len(keyList))
	for idx, key := range keyList {
		prefixKeyList[idx] = fmt.Sprintf(r.keyPrefix, key)
	}
	err := r.redisPipeDel(ctx, prefixKeyList)
	if err != nil {
		component.Logger.Errorf(ctx, "redis set key failed", zap.Error(err), zap.Any("key", prefixKeyList))
	}

	retMap := make(map[string]SetCacheResult)
	for i := 0; i < len(keyList); i++ {
		retMap[keyList[i]] = SetCacheResult{
			ErrHelper: helper.ErrHelper{
				Err: err,
			},
			HandleErrStrategy: r.opts.strategy,
		}
	}
	return retMap
}

func (r *RedisCache) ClearCache(ctx context.Context, key string) ClearCacheResult {
	prefixKey := fmt.Sprintf(r.keyPrefix, key)
	_, err := r.conn.Del(ctx, prefixKey).Result()
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

func (r *RedisCache) BatchClearCache(ctx context.Context, keyList []string) map[string]ClearCacheResult {
	prefixKeyList := make([]string, len(keyList))
	for idx, key := range keyList {
		prefixKeyList[idx] = fmt.Sprintf(r.keyPrefix, key)
	}
	err := r.redisPipeDel(ctx, prefixKeyList)
	if err != nil {
		component.Logger.Errorf(ctx, "redis clear key failed", zap.Error(err), zap.Any("key", prefixKeyList))
	}

	retMap := make(map[string]ClearCacheResult)
	for i := 0; i < len(keyList); i++ {
		retMap[keyList[i]] = ClearCacheResult{
			ErrHelper: helper.ErrHelper{
				Err: err,
			},
			HandleErrStrategy: r.opts.strategy,
		}
	}
	return retMap
}

func (r *RedisCache) RetryGetFromCache(ctx context.Context, key string) GetCacheResult {
	//todo implement
	return GetCacheResult{
		HandleErrStrategy: HandleErrStrategyBreak,
	}
}

func (r *RedisCache) RetrySetCache(ctx context.Context, key string, val string) SetCacheResult {
	//todo implement
	return SetCacheResult{
		HandleErrStrategy: HandleErrStrategyBreak,
	}
}

func (r *RedisCache) RetryClearCache(ctx context.Context, key string) ClearCacheResult {
	//todo implement
	return ClearCacheResult{
		HandleErrStrategy: HandleErrStrategyBreak,
	}
}

func (r *RedisCache) getFromRedis(c context.Context, key string) redisGetResult {
	prefixKey := fmt.Sprintf(r.keyPrefix, key)
	cacheVal := redisGetResult{
		Exist: false,
	}

	v, err := r.conn.Get(c, prefixKey)
	if err == nil && !strings.HasPrefix(v, fmt.Sprintf("%s@", r.opts.tokenPrefix)) {
		cacheVal.Status = RedisCacheStatusOK
		cacheVal.Exist = v != ""
		cacheVal.Value = v
		return cacheVal
	}
	token := r.generateRedisToken()
	temp, err := r.redisCas(c, prefixKey, "", token)
	if err != nil {
		cacheVal.Err = err
		cacheVal.Status = RedisCacheStatusOK
		component.Logger.Errorf(c, "redis get value invalid", zap.Error(err), zap.String("key", prefixKey), zap.Any("value", temp))
		return cacheVal
	}

	if tempString, ok := temp.(string); ok {
		//返回空值 说明当前拿到了token
		//返回token 说明当前已经有请求在走DB
		//返回非token 说明有缓存值
		switch true {
		case tempString == "":
			cacheVal.Token = token
			cacheVal.Status = RedisCacheStatusDoReadDB
		case strings.HasPrefix(tempString, fmt.Sprintf("%s@", r.opts.tokenPrefix)):
			//判断超时
			splitArr := strings.Split(tempString, "@")
			if len(splitArr) != 3 || splitArr[0] != r.opts.tokenPrefix {
				cacheVal.Err = errors.New("token invalid:" + tempString)
				cacheVal.Status = RedisCacheStatusOK
				return cacheVal
			}
			expireTime, err := strconv.ParseInt(splitArr[2], 10, 64)
			if err != nil {
				cacheVal.Err = errors.New("token invalid:" + tempString)
				return cacheVal
			}

			if expireTime < time.Now().Unix() {
				//超时 删掉重试
				if err := r.clearCacheWithToken(c, key, tempString); err != nil {
					cacheVal.Err = err
					cacheVal.Status = RedisCacheStatusOK
					return cacheVal
				}
			}
			//不考虑极端情况 超时就覆盖
			//cacheVal.Token = token
			cacheVal.Status = RedisCacheStatusWait
		default:
			cacheVal.Exist = true
			cacheVal.Status = RedisCacheStatusOK
			cacheVal.Value = tempString
		}
	}

	return cacheVal
}

func (r *RedisCache) batchGetFromRedis(c context.Context, keyList []string) map[string]redisGetResult {
	cacheValList := make(map[string]redisGetResult)
	batchKeyList := make([]string, 0, len(keyList))
	batchTokenList := make(map[string]string)
	batchCheckValList := make(map[string]string)

	prefixKeyList := make([]string, 0, len(keyList))
	prefixMap := make(map[string]string)
	for _, key := range keyList {
		prefixKey := fmt.Sprintf(r.keyPrefix, key)
		prefixKeyList = append(prefixKeyList, prefixKey)
		prefixMap[prefixKey] = key
	}
	tempList, err := r.redisPipeGet(c, prefixKeyList)
	if err != nil {
		component.Logger.Errorf(c, "redis redisPipe invalid", zap.Error(err), zap.Any("key", prefixKeyList))
		for _, key := range keyList {
			ret := redisGetResult{}
			ret.Err = err
			cacheValList[key] = ret
		}
		return cacheValList
	}
	for idx, prefixKey := range prefixKeyList {
		if val, ok := tempList[prefixKey]; ok {
			if valString, ok := val.(string); ok {
				if !strings.HasPrefix(valString, fmt.Sprintf("%s@", r.opts.tokenPrefix)) {
					cacheVal := redisGetResult{}
					cacheVal.Exist = valString != ""
					cacheVal.Value = valString
					cacheVal.Status = RedisCacheStatusOK
					cacheValList[keyList[idx]] = cacheVal
					continue
				}
			}
		}
		batchKeyList = append(batchKeyList, prefixKey)
		batchTokenList[prefixKey] = r.generateRedisToken()
		batchCheckValList[prefixKey] = ""
	}

	if len(batchKeyList) > 0 {
		tempList, err := r.redisCasPipe(c, batchKeyList, batchCheckValList, batchTokenList)
		if err != nil {
			component.Logger.Errorf(c, "redis redisCasPipe invalid", zap.Error(err), zap.Any("key", keyList))
			for _, prefixKey := range batchKeyList {
				ret := redisGetResult{}
				ret.Err = err
				cacheValList[prefixMap[prefixKey]] = ret
			}
			return cacheValList
		}

		for _, prefixKey := range batchKeyList {
			cacheVal := redisGetResult{
				Exist: false,
			}
			temp := tempList[prefixKey]
			if tempString, ok := temp.(string); ok {
				//返回空值 说明当前拿到了token
				//返回token 说明当前已经有请求在走DB
				//返回非token 说明有缓存值
				switch true {
				case tempString == "":
					cacheVal.Token = batchTokenList[prefixKey]
					cacheVal.Status = RedisCacheStatusDoReadDB
				case strings.HasPrefix(tempString, fmt.Sprintf("%s@", r.opts.tokenPrefix)):
					//判断超时
					splitArr := strings.Split(tempString, "@")
					for {
						if len(splitArr) != 3 || splitArr[0] != r.opts.tokenPrefix {
							component.Logger.Errorf(c, "token invalid", zap.Error(err), zap.String("key", prefixKey), zap.Any("value", tempString))
							cacheVal.Err = errors.New("token invalid:" + tempString)
							cacheVal.Status = RedisCacheStatusOK
							break
						}
						expireTime, err := strconv.ParseInt(splitArr[2], 10, 64)
						if err != nil {
							component.Logger.Errorf(c, "token invalid", zap.Error(err), zap.String("key", prefixKey), zap.Any("value", tempString))
							cacheVal.Err = errors.New("token invalid:" + tempString)
							cacheVal.Status = RedisCacheStatusOK
							break
						}

						if expireTime < time.Now().Unix() {
							//超时 删掉重试
							if err := r.clearCacheWithToken(c, prefixMap[prefixKey], tempString); err != nil {
								component.Logger.Errorf(c, "clear cache err", zap.Error(err), zap.String("key", prefixKey), zap.Any("value", tempString))
								cacheVal.Err = err
								cacheVal.Status = RedisCacheStatusOK
								break
							}
						}
						//不考虑极端情况 超时就覆盖
						//cacheVal.Token = token
						cacheVal.Status = RedisCacheStatusWait
						break
					}

				default:
					cacheVal.Exist = true
					cacheVal.Status = RedisCacheStatusOK
					cacheVal.Value = tempString
				}

			}
			cacheValList[prefixMap[prefixKey]] = cacheVal
		}
	}

	return cacheValList
}

func (r *RedisCache) handleGetFromRedis(c context.Context, isBatch bool, handleMap map[string]redisGetResult, waitingLoop int) map[string]GetCacheResult {
	//从缓存找
	waitingList := make([]string, 0, len(handleMap))
	fromNoCacheList := make([]string, 0, len(handleMap))
	retMap := make(map[string]GetCacheResult)
	for key, result := range handleMap {
		ret := GetCacheResult{}
		ret.HandleErrStrategy = r.opts.strategy
		if !result.IsSuccess() {
			ret.Err = result.Err
			ret.Exist = false
			retMap[key] = ret
			continue
		}
		switch result.Status {
		case RedisCacheStatusOK:
			ret.Exist = result.Exist
			ret.Value = result.Value
			retMap[key] = ret
		case RedisCacheStatusWait:
			waitingList = append(waitingList, key)
		case RedisCacheStatusDoReadDB:
			//无缓存直接走 DB
			fromNoCacheList = append(fromNoCacheList, key)
		}
	}

	if waitingLoop >= r.opts.maxWaitingLoop {
		//todo  加配置项 可配置等待超时是走db  还是返回异常
		component.Logger.Warn(c, "baseBatchGet waiting loop over", zap.Any("waitingList", waitingList))
		fromNoCacheList = append(fromNoCacheList, waitingList...)
		waitingList = make([]string, 0, 0)
	}

	//未拿到缓存且拿到读DB权限 去回写
	if len(fromNoCacheList) > 0 {
		var fromNoCacheVal = make(map[string]string)
		var err error
		if isBatch {
			fromNoCacheVal, err = r.batchFn(c, fromNoCacheList)
		} else {
			v, errInner := r.fn(c, fromNoCacheList[0])
			err = errInner
			fromNoCacheVal[fromNoCacheList[0]] = v
		}
		for _, key := range fromNoCacheList {
			ret := GetCacheResult{}
			ret.HandleErrStrategy = r.opts.strategy
			if err != nil {
				component.Logger.Error(c, "baseBatchGet get no cache err", zap.Error(err))
				//删了给别人写
				ret.Err = err
				err = r.clearCacheWithToken(c, key, handleMap[key].Token)
				if err != nil {
					component.Logger.Error(c, "baseBatchGet clear cache err", zap.Error(err))
					ret.Err = errors.Join(ret.Err, err)
				}
				retMap[key] = ret
				continue
			}
			//回写缓存
			if err := r.setCacheWithToken(c, key, handleMap[key].Token, fromNoCacheVal[key]); err != nil {
				component.Logger.Error(c, "baseBatchGet set cache err", zap.Error(err))
				ret.Err = errors.Join(ret.Err, err)
			}
			ret.Exist = fromNoCacheVal[key] != ""
			ret.Value = fromNoCacheVal[key]
			retMap[key] = ret

		}
	}

	//循环处理waiting
	if len(waitingList) > 0 {
		waitingLoop++
		time.Sleep(time.Millisecond * time.Duration(float64(20)+math.Pow(10, float64(waitingLoop-1))))
		var waitingValList = make(map[string]redisGetResult)
		if isBatch {
			waitingValList = r.batchGetFromRedis(c, waitingList)
		} else {
			ret := r.getFromRedis(c, waitingList[0])
			waitingValList[waitingList[0]] = ret
		}
		retMapInner := r.handleGetFromRedis(c, isBatch, waitingValList, waitingLoop)
		for key, val := range retMapInner {
			retMap[key] = val
		}
	}

	return retMap
}

func (r *RedisCache) setCacheWithToken(c context.Context, key, token, value string) error {
	key = fmt.Sprintf(r.keyPrefix, key)
	_, err := r.redisCas(c, key, token, value)
	if err != nil {
		component.Logger.Errorf(c, "set value from redis cache with token invalid (%v)", zap.String("key", key))
	}
	return err

}
func (r *RedisCache) clearCacheWithToken(c context.Context, key string, token string) error {
	key = fmt.Sprintf(r.keyPrefix, key)
	_, err := r.redisCas(c, key, token, "")

	if err != nil {
		component.Logger.Errorf(c, "del value from redis cache with token invalid (%v)", zap.String("key", key))
		return err
	}

	return nil
}

func (r *RedisCache) generateRedisToken() string {
	//prefix@随机数@有效时间
	milliTime := time.Now().UnixMilli()
	expireTime := milliTime/1000 + milliTime%1000
	return fmt.Sprintf("%s@%s%d@%d", r.opts.tokenPrefix, uuid.NewV4().String(), milliTime, expireTime)
}

func (r *RedisCache) redisCas(c context.Context, key string, checkVal string, setVal string) (interface{}, error) {
	//为了避免脏写
	//A -> 读DB (耗时很长) -> 写redis
	//B -> 修改数据 -> 删除 redis
	//如果A读 DB 耗时很长  可能把B修改前数据回写redis  造成历史数据回写
	//因此加token

	script := `local current = redis.call('get',KEYS[1]);
               if not current then
                   current = ''
				end
	           if current == ARGV[1] then 
			       redis.call('setex', KEYS[1], ARGV[3], ARGV[2])
                   return current
               else
                   return current
               end`

	val, err := r.conn.Eval(c, script, []string{key}, checkVal, setVal, r.opts.expireTime)
	if checkVal == "" && err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (r *RedisCache) redisCasPipe(c context.Context, keyList []string, checkValList map[string]string, setValList map[string]string) (map[string]interface{}, error) {
	//为了避免脏写
	//A -> 读DB (耗时很长) -> 写redis
	//B -> 修改数据 -> 删除 redis
	//如果A读 DB 耗时很长  可能把B修改前数据回写redis  造成历史数据回写
	//因此加token
	valList := make(map[string]interface{})
	keyListCount := len(keyList)
	pipeCount := 0
	pipe := r.conn.Pipeline()
	for _, key := range keyList {
		checkVal := checkValList[key]
		setVal := setValList[key]

		script := `local current = redis.call('get',KEYS[1]);
               if not current then
                   current = ''
				end
	           if current == ARGV[1] then 
			       redis.call('setex', KEYS[1], ARGV[3], ARGV[2])
                   return current
               else
                   return current
               end`

		pipe.Eval(c, script, []string{key}, checkVal, setVal, r.opts.expireTime)
		pipeCount++
		if pipeCount%1000 == 0 || pipeCount == keyListCount {
			cmdList, err := pipe.Exec(c)
			if err != nil {
				return valList, err
			}
			for _, cmd := range cmdList {
				if len(cmd.Args()) < 3 {
					return valList, errors.New(fmt.Sprintf("redis exec err %v", zap.Any("args", cmd.Args())))
				}
				key, ok := cmd.Args()[3].(string)
				if !ok {
					return valList, errors.New(fmt.Sprintf("redis exec err %v", zap.Any("args", cmd.Args())))
				}

				val, err := cmd.Result()
				if err != nil {
					return valList, err
				}

				valList[key] = val
			}
		}

	}

	return valList, nil
}

func (r *RedisCache) redisPipeGet(c context.Context, keyList []string) (map[string]interface{}, error) {
	//为了避免脏写
	//A -> 读DB (耗时很长) -> 写redis
	//B -> 修改数据 -> 删除 redis
	//如果A读 DB 耗时很长  可能把B修改前数据回写redis  造成历史数据回写
	//因此加token
	keyListCount := len(keyList)
	valList := make(map[string]interface{}, keyListCount)
	pipeCount := 0
	pipe := r.conn.Pipeline()
	for _, key := range keyList {
		script := `return redis.call('get',KEYS[1]);`
		pipe.Eval(c, script, []string{key})
		pipeCount++
		if pipeCount%1000 == 0 || pipeCount == keyListCount {
			cmdList, err := pipe.Exec(c)
			if err != nil && err != redis.Nil {
				return valList, err
			}
			for _, cmd := range cmdList {
				if len(cmd.Args()) < 3 {
					return valList, errors.New(fmt.Sprintf("redis exec err %v", zap.Any("args", cmd.Args())))
				}
				key, ok := cmd.Args()[3].(string)
				if !ok {
					return valList, errors.New(fmt.Sprintf("redis exec err %v", zap.Any("args", cmd.Args())))
				}

				val, err := cmd.Result()
				if err == redis.Nil {
					continue
				}
				if err != nil {
					return valList, err
				}

				valList[key] = val
			}
		}

	}
	return valList, nil
}

func (r *RedisCache) redisPipeDel(c context.Context, keyList []string) error {
	keyListCount := len(keyList)
	pipeCount := 0
	pipe := r.conn.Pipeline()
	for _, key := range keyList {
		script := `return redis.call('del',KEYS[1]);`
		pipe.Eval(c, script, []string{key})
		pipeCount++
		if pipeCount%1000 == 0 || pipeCount == keyListCount {
			cmdList, err := pipe.Exec(c)
			if err != nil {
				return err
			}
			for _, cmd := range cmdList {
				if _, err := cmd.Result(); err != nil {
					return err
				}
			}
		}

	}
	return nil
}
