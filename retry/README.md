### 重试组件

一个灵活的 Go 语言重试组件，提供多种重试策略和配置选项。

### 特征

1. 支持多种重试策略：
   - 最大重试次数策略 (MaxTimes)
   - 直到成功策略 (UntilSuccess)
2. 支持指数退避算法
3. 支持自定义重试间隔
4. 支持最大延迟时间限制
5. 内置抖动机制，避免惊群效应
6. 支持带返回值的重试函数

### 安装

```bash
go get github.com/graymonster0927/component
```

### 使用

#### 1. 最大重试次数策略

```go
// 创建重试实例
retry, err := GetRetryHelper(RetryMaxTimes, 
    WithMaxTimesMaxTimes(3),                    // 最大重试3次
    WithMaxTimesRetryTimeout(100*time.Millisecond), // 重试间隔100ms
    WithMaxTimesMaxDelay(500*time.Millisecond),     // 最大延迟500ms
    WithMaxTimesIsExponential(true),                // 使用指数退避
)

// 执行无返回值的重试
err = retry.DoRetry(func() error {
    // 你的业务逻辑
    return nil
})

// 执行有返回值的重试
result, err := retry.DoRetryReturn(func() (interface{}, error) {
    // 你的业务逻辑
    return "success", nil
})
```

#### 2. 直到成功策略

```go
// 创建重试实例
retry, err := GetRetryHelper(RetryUntilSuccess,
    WithUntilSuccessMaxTimes(3),
    WithUntilSuccessRetryTimeout(100*time.Millisecond),
    WithUntilSuccessMaxDelay(500*time.Millisecond),
    WithUntilSuccessIsExponential(true),
)

// 注册重试方法
SetUntilSuccessRetryMethod(retry, "task1", func(params ...any) error {
	v1 := params[0].(string)
	v2 := params[1].(int)
	fmt.Println(v1, v2)
    // 你的业务逻辑
    return nil
})

// 执行重试
err = retry.DoRetryWithParams("task1", "justtest", 666)
```

使用场景:

* 重试

### 配置说明

* MaxTimes: 最大重试次数
* RetryTimeout: 重试间隔时间
* MaxDelay: 最大延迟时间
* IsExponential: 是否使用指数退避算法

### TODO
* 支持更多重试策略
* 支持自定义重试条件
* 支持重试事件回调

### 贡献

欢迎提交问题（issues）或请求（pull requests）以帮助改进该库。
