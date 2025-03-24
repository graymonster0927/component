## 链式缓存

CacheChain 是一个Go语言编写的链式缓存管理库，提供灵活且可扩展的接口，用于管理多级缓存层。通过该库，开发者可以轻松地与多种缓存后端（如内存缓存、Redis等）进行交互，执行缓存的 Get、Set、Clear 等操作，并具备错误处理和重试机制。

### 特征
 * 链式缓存：支持多层缓存链。当一个缓存层获取数据失败时，会自动尝试从下一个缓存层获取。
 * 错误处理策略：支持多种错误处理策略（Continue、Break、Retry），开发者可以根据不同的缓存层配置不同的处理方式。
 * 灵活扩展：可轻松集成多个缓存后端，并根据需要添加新的缓存层。
 * 已实现redis缓存, redis缓存支持pipeline批量获取数据, 实现旁路缓存, 保证缓存和DB一致性

### 场景
 * 单独用redis旁路缓存, 当做分布式缓存的场景
 * memory + redis, 对redis前面加一层内存缓存, 适合qps非常高, 且能够接受数据适当延迟修改的场景


### 安装

```
go get github.com/graymonster0927/component
```
### 使用

创建缓存链实例
```
    chain := cachechain.NewCacheChain()
    //创建一个 redis 缓存
    redisCache := cache.NewRedisCache(
        //可以实现自己的redis连接
	cache.WithRedisConn(&RedisV8{}),
	cache.WithTokenPrefix("cachechain:servicename"),
    )
	
    // 添加多个缓存层（如内存缓存、Redis缓存等）
    chain.WithCache(redisCache)

```
获取数据
```
    chain := helpers.GetRedisCacheChain()
    //设置缓存前缀 比如 servicename:xxx:%d
    chain.SetKeyPrefix(components.ApiResponseWhiteCacheKey)

    fn := func() (username, error) {
        //数据库操作 比如根据电话号码拿用户名
	 return XXXXX()
    }
    chain.SetFnGetNoCache(func(cCtx context.Context, key string) (string, error) {
        return fn()
    })
    telephone := "12345678901"
    getRet := chain.Get(ctx, telephone)
    if !getRet.IsSuccess() {
        //缓存异常 直接不走缓存拿数据
        return fn()
        //....
    }

    if v.Exist {
        //缓存存在 处理 返回数据
        //xxx
    }
```

写缓存
```
    chain := helpers.GetRedisCacheChain()
    chain.SetKeyPrefix(components.ApiResponseWhiteCacheKey)
    keyList := make([]string, len(relID))
    valList := make([]string, len(relID), len(relID))
    for i, v := range relID {
        keyList[i] = fmt.Sprintf("%d", v)
    }
    setRet := chain.BatchSet(ctx, keyList, valList)
    if !setRet.IsSuccess() {
        //重试
        //返回异常
    }
```

批量获取数据
```
    chain := helpers.GetRedisCacheChain()
    chain.SetKeyPrefix(components.ApiResponseWhiteCacheKey)

    fn := func() ([]User, error) {
        //数据库批量操作 比如根据电话号码批量拿用户列表
        return XXXXX()
    }
    chain.SetFnBatchGetNoCache(func(cCtx context.Context, keyList []string) (map[string]string, error) {
    ret := make(map[string]string)
    list, err := fn()
    for _, v := range list {
        ret[v.telephone] = v.name
    }
        return ret, err
    })
    telephoneList := []string{"12345678901", "12345678902"}
    getRetMap := chain.BatchGet(ctx, telephoneList)
    for _, v := range getRetMap {
        if !v.IsSuccess() {
        //缓存异常 直接不走缓存拿数据
        list, err := fn()
        //....
        }

        if v.Exist {
            //缓存存在 处理 返回数据
            //xxx
        }
    }
```

批量写缓存
```
    chain := helpers.GetRedisCacheChain()
    chain.SetKeyPrefix(components.ApiResponseWhiteCacheKey)
    keyList := make([]string, len(relID))
    valList := make([]string, len(relID), len(relID))
    for i, v := range relID {
        keyList[i] = fmt.Sprintf("%d", v)
    }
    setRetMap := chain.Set(ctx, keyList, valList)
    for _, v := range setRetMap {
        if !v.IsSuccess() {
            //重试
            //返回异常
        }
    }
```

### 错误处理策略
每个缓存层都可以设置错误处理策略，以决定在遇到错误时的行为：

* HandleErrStrategyContinue：忽略错误，继续检查下一个缓存层。
* HandleErrStrategyBreak：遇到错误时停止并返回错误。
* HandleErrStrategyRetry：在遇到错误时重试操作。

### TODO
* 支持按类型配置GetNoCache函数, 这样全局可使用单实例缓存链
* 错误处理策略支持回滚策略
* 实现内存缓存,文件缓存等
* 完善参数校验

### 贡献
欢迎提交问题（issues）或请求（pull requests）以帮助改进该库。
