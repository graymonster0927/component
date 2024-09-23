package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/graymonster0927/component"
	"github.com/graymonster0927/component/cachechain/helper"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"math"
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
	expireTime int
}

func WithExpireTime(expireTime int) RedisCacheOption {
	return func(o *options) {
		o.expireTime = expireTime
	}
}

type RedisCache struct {
	opts    options
	fn      func(key string) (string, error)
	batchFn func(keyList []string) (map[string]string, error)
}

func NewRedisCache(opts ...RedisCacheOption) *RedisCache {
	// 创建一个默认的options
	op := options{
		expireTime: 7 * 86400,
	}
	// 调用动态传入的参数进行设置值
	for _, option := range opts {
		option(&op)
	}
	return &RedisCache{
		opts: op,
	}
}

func (r *RedisCache) SetFnGetNoCache(fn func(key string) (string, error)) {
	r.fn = fn
}

func (r *RedisCache) SetFnBatchGetNoCache(fn func(keyList []string) (map[string]string, error)) {
	r.batchFn = fn
}

func (r *RedisCache) GetFromCache(c context.Context, key string) getCacheResult {

}

func (r *RedisCache) getFromRedis(c context.Context, key string) redisGetResult {
	cacheVal := redisGetResult{
		Exist: false,
	}

	token := r.generateRedisToken()
	temp, err := r.redisCas(c, key, "", token)
	if err != nil {
		cacheVal.Err = err
		component.Logger.Errorf(c, "redis get value invalid", zap.Error(err), zap.String("key", key), zap.Any("value", temp))
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
		case strings.HasPrefix(tempString, "eeo-uc-token@"):
			//判断超时
			splitArr := strings.Split(tempString, "@")
			if len(splitArr) != 3 || splitArr[0] != "eeo-uc-token" {
				cacheVal.Err = errors.New("token invalid:" + tempString)
				return cacheVal
			}
			expireTime, err := strconv.ParseInt(splitArr[2], 10, 64)
			if err != nil {
				cacheVal.Err = errors.New("token invalid:" + tempString)
				return cacheVal
			}

			if expireTime < time.Now().Unix() {
				//超时 删掉重试
				if err := r.ClearCache(c, key, tempString); err != nil {
					cacheVal.Err = err
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

func (r *RedisCache) handleGetFromRedis(c *context.Context, cacheKey string, searchList []string, fnGetFromDB func(searchList []string) (map[string]string, error), waitingLoop uint8) (map[string]string, error) {
	//从缓存找
	waitingList := make([]string, 0, len(searchList))
	excludeList := make([]string, 0, len(searchList))
	cachedValList := make(map[string]string)
	keyList := make([]string, 0, len(searchList))
	for _, search := range searchList {
		key := fmt.Sprintf(cacheKey, search)
		keyList = append(keyList, key)
	}

	//拿cache
	modelCacheList, err := modelCacheHelper.BatchGetFromCache(c, keyList)
	if err != nil {
		return nil, err
	}
	for _, search := range searchList {
		key := fmt.Sprintf(cacheKey, search)
		modelCache := modelCacheList[key]
		switch modelCache.Status {
		case cache.ModelCacheStatusOK:
			cachedValList[key] = modelCache.Value
		case cache.ModelCacheStatusWait:
			waitingList = append(waitingList, search)
		case cache.ModelCacheStatusDoReadDB:
			//无缓存直接走 DB
			excludeList = append(excludeList, search)
		}
	}

	//都走缓存
	if len(cachedValList) == len(searchList) {
		return cachedValList, nil
	}

	if waitingLoop >= WaitingLoopMax {
		library.WithUUIDLogger(c).Warn("baseBatchGet from db warning", zap.Any("key", cacheKey), zap.Any("waitingList", waitingList))
		excludeList = append(excludeList, waitingList...)
		waitingList = make([]string, 0, 0)
	}

	//未拿到缓存且拿到读DB权限 去回写
	if len(excludeList) > 0 {
		excludeValList, err := fnGetFromDB(excludeList)
		if err != nil {
			return cachedValList, err
		}
		for key, val := range excludeValList {
			if modelCacheList[key].Status == cache.ModelCacheStatusDoReadDB {
				if err := modelCacheHelper.SetCache(c, key, cache.ModelCache{
					Token: modelCacheList[key].Token,
					Value: val,
				}); err != nil {
					library.WithUUIDLogger(c).Error("baseBatchGet set cache err", zap.Error(err))
				}
			}
			cachedValList[key] = val
		}
	}

	//循环处理waiting
	if len(waitingList) > 0 {
		waitingLoop++
		time.Sleep(time.Millisecond * time.Duration(float64(20)+math.Pow(10, float64(waitingLoop-1))))
		waitingValList, err := baseBatchGet(c, cacheKey, waitingList, fnGetFromDB, waitingLoop)
		if err != nil {
			return cachedValList, err
		}
		for key, val := range waitingValList {
			cachedValList[key] = val
		}
	}

	return cachedValList, nil

}

//func (m *RedisCache) BatchGetFromCache(c *gin.Context, keyList []string) (map[string]ModelCache, error) {
//	cacheValList := make(map[string]ModelCache)
//	batchKeyList := make([]string, 0, len(keyList))
//	batchTokenList := make(map[string]string)
//	batchCheckValList := make(map[string]string)
//	//for _, key := range keyList {
//	//	cacheVal := ModelCache{
//	//		Exist:false,
//	//	}
//	//	if m.opts.memoryCacheEnable {
//	//		if val, ok := GetMemoryCacheInstance().GetMemoryCache(key); ok {
//	//			if valString, ok := val.(string); ok {
//	//				cacheVal.Exist = true
//	//				cacheVal.Value = valString
//	//				cacheVal.Status = ModelCacheStatusOK
//	//				cacheValList[key] = cacheVal
//	//				continue
//	//			}
//	//			library.WithUUIDLogger(c).Error("get value from memory cache invalid", zap.String("key", key), zap.Any("value", val))
//	//		}
//	//	}
//
//	tempList, err := m.redisPipe(c, keyList)
//	if err != nil {
//		library.WithUUIDLogger(c).Error("get value from redis cache invalid", zap.Error(err), zap.Any("key", keyList), zap.Any("value", tempList))
//		return cacheValList, err
//	}
//	for _, key := range keyList {
//
//		if val, ok := tempList[key]; ok {
//			if valString, ok := val.(string); ok {
//				if valString != "" && !strings.HasPrefix(valString, "eeo-uc-token@") {
//					cacheVal := ModelCache{
//						Exist: false,
//					}
//					cacheVal.Exist = true
//					cacheVal.Value = valString
//					cacheVal.Status = ModelCacheStatusOK
//					cacheValList[key] = cacheVal
//					continue
//				}
//			}
//		}
//		batchKeyList = append(batchKeyList, key)
//		batchTokenList[key] = m.generateRedisToken()
//		batchCheckValList[key] = ""
//	}
//
//	if len(batchKeyList) > 0 {
//		tempList, err := m.redisCasPipe(c, batchKeyList, batchCheckValList, batchTokenList)
//		if err != nil {
//			library.WithUUIDLogger(c).Error("get value from redis cache invalid", zap.Error(err), zap.Any("key", batchKeyList), zap.Any("value", tempList))
//			return cacheValList, err
//		}
//		for _, key := range batchKeyList {
//			cacheVal := ModelCache{
//				Exist: false,
//			}
//			temp := tempList[key]
//			if tempString, ok := temp.(string); ok {
//				//返回空值 说明当前拿到了token
//				//返回token 说明当前已经有请求在走DB
//				//返回非token 说明有缓存值
//				switch true {
//				case tempString == "":
//					cacheVal.Token = batchTokenList[key]
//					cacheVal.Status = ModelCacheStatusDoReadDB
//				case strings.HasPrefix(tempString, "eeo-uc-token@"):
//					//判断超时
//					splitArr := strings.Split(tempString, "@")
//					if len(splitArr) != 3 || splitArr[0] != "eeo-uc-token" {
//						return cacheValList, errors.New("token invalid:" + tempString)
//					}
//					expireTime, err := strconv.ParseInt(splitArr[2], 10, 64)
//					if err != nil {
//						return cacheValList, errors.New("token invalid:" + tempString)
//					}
//
//					if expireTime < time.Now().Unix() {
//						//超时 删掉重试
//						if err := m.ClearCache(c, key, tempString); err != nil {
//							return cacheValList, err
//						}
//					}
//					//不考虑极端情况 超时就覆盖
//					//cacheVal.Token = token
//					cacheVal.Status = ModelCacheStatusWait
//				default:
//					cacheVal.Exist = true
//					cacheVal.Status = ModelCacheStatusOK
//					cacheVal.Value = tempString
//					//如果内存缓存允许 回写
//					//if m.opts.memoryCacheEnable {
//					//todo 单飞
//					//	GetMemoryCacheInstance().SetMemoryCache(key, cacheVal.Value, time.Second * 5)
//					//}
//
//				}
//
//			}
//			cacheValList[key] = cacheVal
//		}
//	}
//
//	return cacheValList, nil
//}

//	func (m *ModelCacheHelper)SetCache(c *gin.Context, key string, cacheVal ModelCache) error {
//		_, err := m.redisCas(c, key, cacheVal.Token, cacheVal.Value)
//		if err != nil {
//			library.WithUUIDLogger(c).Error("set value from redis cache invalid", zap.String("key", key), zap.Any("value", cacheVal.Value))
//			return err
//		}
//
//		//成功设置
//		//if currentToken == cacheVal.Token {
//		//	if m.opts.memoryCacheEnable {
//		//		GetMemoryCacheInstance().SetMemoryCache(key, cacheVal.Value, time.Second * 5)
//		//	}
//		//}
//
//		return nil
//	}
func (r *RedisCache) ClearCache(c context.Context, key string, token string) error {
	_, err := r.redisCas(c, key, token, "")

	if err != nil {
		component.Logger.Errorf(c, "del value from redis cache invalid", zap.String("key", key))
		return err
	}

	return nil
}

func (r *RedisCache) generateRedisToken() string {
	//随机数-有效时间
	//return fmt.Sprintf("eeo-uc-token@%s@%d", rand.Int63n(1000000000000000), time.Now().Unix() + utils.Minute)
	return fmt.Sprintf("eeo-uc-token@%s@%d", uuid.NewV4().String(), time.Now().Unix()+utils.Minute)
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

	val, err := database.RedisModelCache.Eval(c, script, []string{key}, checkVal, setVal, r.opts.expireTime).Result()
	if checkVal == "" && err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (r *RedisCache) redisCasPipe(c *gin.Context, keyList []string, checkValList map[string]string, setValList map[string]string) (map[string]interface{}, error) {
	//为了避免脏写
	//A -> 读DB (耗时很长) -> 写redis
	//B -> 修改数据 -> 删除 redis
	//如果A读 DB 耗时很长  可能把B修改前数据回写redis  造成历史数据回写
	//因此加token
	valList := make(map[string]interface{})
	keyListCount := len(keyList)
	pipeCount := 0
	pipe := database.RedisModelCache.Pipeline()
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

		pipe.Eval(c, script, []string{key}, checkVal, setVal, m.opts.redisExpireTime)
		pipeCount++
		if pipeCount%1000 == 0 || pipeCount == keyListCount {
			cmdList, err := pipe.Exec(c)
			if err != nil {
				return valList, err
			}
			for _, cmd := range cmdList {
				cmd := cmd.(*redis.Cmd)
				if len(cmd.Args()) < 3 {
					return valList, errors.Errorf("redis exec err %s", zap.Any("args", cmd.Args()))
				}
				key, ok := cmd.Args()[3].(string)
				if !ok {
					return valList, errors.Errorf("redis exec err %s", zap.Any("args", cmd.Args()))
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

func (r *RedisCache) redisPipe(c *gin.Context, keyList []string) (map[string]interface{}, error) {
	//为了避免脏写
	//A -> 读DB (耗时很长) -> 写redis
	//B -> 修改数据 -> 删除 redis
	//如果A读 DB 耗时很长  可能把B修改前数据回写redis  造成历史数据回写
	//因此加token
	keyListCount := len(keyList)
	valList := make(map[string]interface{}, keyListCount)
	pipeCount := 0
	pipe := database.RedisModelCache.Pipeline()
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
				cmd := cmd.(*redis.Cmd)
				if len(cmd.Args()) < 3 {
					return valList, errors.Errorf("redis exec err %s", zap.Any("args", cmd.Args()))
				}
				key, ok := cmd.Args()[3].(string)
				if !ok {
					return valList, errors.Errorf("redis exec err %s", zap.Any("args", cmd.Args()))
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
