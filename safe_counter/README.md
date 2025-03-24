# SafeCounter 全局计数器

`safecounter` 包提供了一个安全的计数器实现，可用于在并发环境中对键值进行计数操作。

## 功能特点

- 协程安全：保证在并发环境中的数据一致性
- 键值计数：支持对 `string` 类型的键进行计数
- 灵活操作：支持单次增加、批量增加、获取计数值、清理等操作

## 安装


```
go get github.com/graymonster0927/component
```

## 使用示例


```go
package main

import (
    "fmt"
    "github.com/graymonster0927/component/safecounter"
)

func main() {
    // 初始化计数器
    counter := safecounter.NewSafeCounter()
    
    // 增加计数
    counter.Inc("200") // 增加 HTTP 200 状态码的计数
    counter.Inc("404") // 增加 HTTP 404 状态码的计数
    counter.IncN("500", 3) // 增加 3 次 HTTP 500 状态码的计数
    
    // 获取所有计数并清空
    counts := counter.All(true)
    
    // 打印结果
    for code, count := range counts {
        fmt.Printf("HTTP %d: %d responses\n", code, count)
    }
    
}
```
