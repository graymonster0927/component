### 事件机制


#### 例子
```
func BootstrapListener() {
	l := listener.GetEventListenerInstance()
	//demo 异步处理 listener 注册
	l.RegisterListener(&listener.DemoAsyncListener{})
	//demo 同步处理 listener 注册
	l.RegisterListener(&listener.DemoListener{})
	
	//每个listener处理event, 会根据event的name对不同的event走不同逻辑处理
	
}

func main() {
	//注册 listener
	BootstrapListener()
	
	//触发 event
	if err := dispatcher.GetDispatcherInstance().Dispatch(&event.DemoEvent{}); err != nil {
		//.....
	}

	//触发 event
	if err := dispatcher.GetDispatcherInstance().Dispatch(&event.DemoAsyncEvent{}); err != nil {
		//....
	}
}
```