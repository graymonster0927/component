package listener

import (
	"github.com/graymonster0927/component/event/event"
	"reflect"
)

type DemoAsyncListener struct {
}

func (dl *DemoAsyncListener) Listen() []event.Interface {
	return []event.Interface{
		&event.DemoAsyncEvent{},
	}
}

func (dl *DemoAsyncListener) Handle(e event.Interface) error {
	return nil

}

func (dl *DemoAsyncListener) HandleAsync(e event.Interface) {
	switch e.Name() {
	case (&event.DemoEvent{}).Name():
		// ii := e.(*event.DemoEvent)
		// fmt.Println(ii.Name())
	}
}

func (dl *DemoAsyncListener) IsErrStopped() bool {
	return false
}

func (dl *DemoAsyncListener) Name() string {
	return reflect.TypeOf(dl).String()
}
