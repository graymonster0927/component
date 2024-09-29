### 状态转换检查

Transfer State Checker 是一个用于管理状态转换关系的库，适用于需要检查状态之间合法转换的场景。

### 特征
1. 简化代码 统一封装

### 安装

```
go get github.com/graymonster0927/component
```
### 使用

```go
    package main
    func Test() {
		//对订单的状态修改判断是否合法
	checker := NewTransferStateChecker()

	var stateTypeOrder StateType = 1
	var stateOrderDefault = 1
	var stateOrderIng = 1
	var stateOrderDone = 1
	var stateOrderClosed = 1
	var stateOrderCancel = 1

	checker.SetStateRel(stateTypeOrder, map[int]map[int]struct{}{
		stateOrderDefault: {
			stateOrderIng: {},
		},
		stateOrderIng: {
			stateOrderDone: {},
			stateOrderClosed: {},
			stateOrderCancel: {},
		},
	})
	
	if checker.Check(stateTypeOrder, stateOrderDefault, stateOrderIng) {
		//update order state
    }
}

```

### 贡献
欢迎提交问题（issues）或请求（pull requests）以帮助改进该库。

