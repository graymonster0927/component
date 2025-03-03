### RESTful API 路径格式化组件

对restful风格的接口扫描的工具, 比如 `/api/v1/users/zhangsan/profile` 会拿到 `/api/v1/users/*/profile`
当前聚合是基于某个模式的 url 出现超过阈值的次数，进行路径格式化。比如阈值配置为3 那么:
`/api/v1/users/zhangsan/profile`
`/api/v1/users/lisi/profile`
`/api/v1/users/wangwu/profile`

上面的路径能被识别到 `/api/v1/users/*/profile`
### 特性

1. 实现 RESTful 模式扫描
2. 支持高并发处理
3. 内置等待队列机制
4. 可配置的并发阈值
5. 支持优雅的可视化输出

### 安装

```bash
go get github.com/graymonster0927/component
```

### 使用

#### 1. 基础使用

```go
import "github.com/graymonster0927/component/restful_formater"

// 获取格式化器实例
formatter := restful_formater.GetFormatter()

// 记录 API 路径
err := formatter.RecordAPI("/api/v1/users/profile")
if err != nil {
    // 处理错误
}

// 打印树形结构
fmt.Println(formatter.String())
```

输出示例：
```
API Tree:
api
└── v1
    └── users
        └── profile
```

#### 2. 配置并发阈值

```go
// 设置并发处理阈值
formatter = restful_formater.WithThreshold(3)

// 设置等待队列长度
formatter = restful_formater.WithWaitingList(1000)
```

#### 3. 扫描 RESTful 模式

```go

// 记录 API 路径
err := formatter.RecordAPI("/api/v1/users/zhangsan/profile")
if err != nil {
// 处理错误
}

err := formatter.RecordAPI("/api/v1/users/zhangsan1/profile")
if err != nil {
// 处理错误
}

err := formatter.RecordAPI("/api/v1/users/zhangsan2/profile")
if err != nil {
// 处理错误
}

// 打印树形结构
fmt.Println(formatter.String())

// 扫描并获取 RESTful 模式
patterns, err := formatter.ScanRestfulPattern()
if err != nil {
    // 处理错误
}

// 打印所有模式
// 输出示例：
// ["api/v1/users/*/profile"]
for _, pattern := range patterns {
    fmt.Println(pattern)
}
```


### 配置说明

* `threshold`: 并发处理阈值，默认为 5
* `waitingList`: 等待队列长度，默认为 10240

### 性能特性

1. 使用 sync.Map 保证并发安全
2. 采用细粒度锁优化性能
3. 支持请求排队和限流
4. 内存友好的树形结构


### TODO

* 扫描到可聚合数据后从树中清除
* 支持自定义输出格式

### 贡献

欢迎提交问题（issues）或请求（pull requests）以帮助改进该库。
