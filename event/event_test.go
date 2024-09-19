package event_test

import (
	"github.com/graymonster0927/component/event/dispatcher"
	"github.com/graymonster0927/component/event/event"
	"github.com/graymonster0927/component/event/listener"
	"testing"
)

func TestEventBootstrap(t *testing.T) {
	//先注册
	listener.GetEventListenerInstance().RegisterListener(&listener.DemoListener{})
	listener.GetEventListenerInstance().RegisterListener(&listener.DemoAsyncListener{})
}
func TestEventDispatch(t *testing.T) {
	err := dispatcher.GetDispatcherInstance().Dispatch(&event.DemoEvent{})
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
}

func TestEventDispatched(t *testing.T) {
	err := dispatcher.GetDispatcherInstance().Dispatch(&event.DemoEvent{})
	if err == nil {
		t.Fail()
	} else {
		t.Log(err.Error())
	}

}

func BenchmarkEventDispatch(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := dispatcher.GetDispatcherInstance().Dispatch(&event.DemoEvent{})
			if err != nil {
				b.Log(err.Error())
				b.Fail()
			}
		}
	})
}

func BenchmarkEventDispatchAsync(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := dispatcher.GetDispatcherInstance().Dispatch(&event.DemoAsyncEvent{})
			if err != nil {
				b.Log(err.Error())
				b.Fail()
			}
		}
	})
}
