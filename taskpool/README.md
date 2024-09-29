### 任务池

### 特征
1. 任务通过 goroutine 并发执行
2. 简化代码 统一封装 
3. 加入协程池 减少gc 通过 [ants](https://github.com/panjf2000/ants) 支持。

### 安装

```
go get github.com/graymonster0927/component
```
### 使用


使用场景:

* 并发执行多个任务, 且需要等待所有结果

> 比如 同一张图 -> 同时 校验A规则/校验B规则/校验C规则 -> 任一个规则失败则不允许图保存
```

const (
	TaskTypeDemo1 TaskType = 1
	TaskTypeDemo2 TaskType = 2
	TaskTypeDemo3 TaskType = 3
)

func main() {
	//获取任务池
	ctx := context.Background()
	taskPool := GetTaskPool(&ctx)
	//设置协程池大小
    taskPool.SetGPoolSize(100)
	//设置任务池的任务处理函数
	taskPool.SetTaskHandler(TaskTypeDemo1, func(ctx *context.Context, params map[string]interface{}) (interface{}, error) {
		fmt.Println("demo1")
		return nil, nil
	})

	say := "my name is a"
	taskPool.SetTaskHandler(TaskTypeDemo2, func(ctx *context.Context, params map[string]interface{}) (interface{}, error) {
		fmt.Println(say)
		return nil, nil
	})

	taskPool.SetTaskHandler(TaskTypeDemo3, func(ctx *context.Context, params map[string]interface{}) (interface{}, error) {
		p1 := params["p1"].(string)
		p2 := params["p2"].(int)
		fmt.Println(p1, p2)
		return nil, nil
	})

	//添加任务
	taskPool.AddTask(TaskTypeDemo1, "id-1", nil)
	taskPool.AddTask(TaskTypeDemo1, "id-2", nil)
	taskPool.AddTask(TaskTypeDemo2, "id-1", nil)

	params := map[string]interface{}{
		"p1": "a",
		"p2": 1,
	}
	taskPool.AddTask(TaskTypeDemo3, "id-1", params)

	if err := taskPool.Start(); err != nil {
		fmt.Println("task exception", err)
		return
	}

	fmt.Println(taskPool.errList)
	fmt.Println(taskPool.retList)

}


```

* 限制并发执行任务数目

> 比如限制一次只能执行500个扫描端口的任务 避免扫描端口过多导致NAT压力过大

```

const (
	TaskTypeDemo3 TaskType = 3
)

func main() {
	//获取任务池
	ctx := context.Background()
	taskPool := GetTaskPool(&ctx)

	//设置任务池的任务处理函数
	taskPool.SetTaskHandler(TaskTypeDemo3, func(ctx *context.Context, params map[string]interface{}) (interface{}, error) {
		p1 := params["p1"].(int)
		fmt.Println(p1+"port is scanning")
		return nil, nil
	})

	//添加任务
    for i := 0; i < 65535; i++ {
        count++
        params := map[string]interface{}{
		    "p1": i,
	    }
        taskPool.AddTask(TaskTypeDemo3, "id-"+strconv.Itoa(i), params)
        if count % 500 == 0 {
            if err := taskPool.Start(); err != nil {
                fmt.Println("task exception", err)
                return
            }
            
            //处理结果
            taskPool.Clear()
        }
    }
    
    if count % 500 > 0 {
        if err := taskPool.Start(); err != nil {
            fmt.Println("task exception", err)
            return
        }
        //处理结果
        taskPool.Clear()
    }

}

```

### TODO
* 支持按任务维度限速

### 贡献
欢迎提交问题（issues）或请求（pull requests）以帮助改进该库。