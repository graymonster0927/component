# 事件机制

本项目实现了一种事件监听器（Listener）和事件调度器（Dispatcher）机制，通过该机制，你可以注册事件监听器，并触发事件进行处理。该机制支持同步和异步的事件处理方式。

## 快速开始

### 1. 监听器注册

在程序启动时，首先需要注册监听器。通过 `BootstrapListener` 方法进行监听器的注册。

```go
package main
func BootstrapListener() {
	
	//该方法中，我们分别注册了异步和同步两种类型的监听器：
	//DemoAsyncListener：用于异步处理事件。
	//DemoListener：用于同步处理事件。
	
    l := listener.GetEventListenerInstance()

    // 异步处理事件的监听器注册
    l.RegisterListener(&listener.DemoAsyncListener{})

    // 同步处理事件的监听器注册
    l.RegisterListener(&listener.DemoListener{})
}
```

### 2. 触发事件
   
在 main 函数中，我们通过事件调度器（Dispatcher）来触发事件。不同的事件可以有不同的监听器进行处理。

```go
package main
func main() {
    // 注册监听器
    BootstrapListener()

	//在下面的代码中，我们触发了两个事件：
	//DemoEvent：这是一个同步事件，会被同步监听器 DemoListener 处理。
	//DemoAsyncEvent：这是一个异步事件，会被异步监听器 DemoAsyncListener 处理。
	
    // 触发同步事件
    if err := dispatcher.GetDispatcherInstance().Dispatch(&event.DemoEvent{}); err != nil {
        // 处理错误
        log.Println("Error dispatching DemoEvent:", err)
    }

    // 触发异步事件
    if err := dispatcher.GetDispatcherInstance().Dispatch(&event.DemoAsyncEvent{}); err != nil {
        // 处理错误
        log.Println("Error dispatching DemoAsyncEvent:", err)
    }
}
```

## 机制说明

### 事件监听器（Listener）
监听器是用来监听并处理不同事件的。你可以为同一类事件注册多个监听器。监听器的处理逻辑可以根据事件的类型或名称进行不同的逻辑处理。

### 事件调度器（Dispatcher）
事件调度器负责将事件分发给相应的监听器。你只需通过 Dispatch 方法触发事件，调度器会自动将事件传递给已注册的监听器进行处理。

### 异步与同步
同步处理：事件触发后，程序会等待事件处理完成后再继续执行后续逻辑。
异步处理：事件触发后，程序不需要等待处理完成，监听器会在后台处理该事件。

### 注意事项
确保在触发事件之前，已经完成监听器的注册。
异步监听器的处理逻辑需要确保线程安全。
在分发事件时，注意捕获并处理错误，避免因事件处理异常导致程序崩溃。

### 扩展
你可以通过实现新的监听器来扩展系统的功能，只需实现对应的接口，并注册到事件监听器即可。以下是一个自定义监听器的示例：

```go
type MyCustomListener struct{}

func (l *MyCustomListener) HandleEvent(e event.Event) {
    // 自定义事件处理逻辑
}
//然后在 BootstrapListener 中注册该监听器：

l.RegisterListener(&MyCustomListener{})
```

## 总结
该事件机制可以方便地实现基于事件驱动的开发模式，支持同步和异步事件的处理，非常适合用于需要监听不同事件并做出相应处理的场景。

### 贡献
欢迎提交问题（issues）或请求（pull requests）以帮助改进该库。