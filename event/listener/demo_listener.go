package listener

import (
	"github.com/graymonster0927/component/event/event"
	"reflect"
)

type DemoListener struct {
}

func (dl *DemoListener) Listen() []event.Interface {
	return []event.Interface{
		&event.DemoEvent{},
	}
}

func (dl *DemoListener) Handle(e event.Interface) error {
	switch e.Name() {
	case (&event.DemoEvent{}).Name():
		// ii := e.(*event.DemoEvent)
		// fmt.Println(ii.Name())
		return nil
	}
	return nil
}

func (dl *DemoListener) HandleAsync(e event.Interface) {
	switch e.Name() {
	case (&event.DemoEvent{}).Name():
		// ii := e.(*event.DemoEvent)
		// fmt.Println(ii.Name())
	}
}

func (dl *DemoListener) IsErrStopped() bool {
	return false
}

func (dl *DemoListener) Name() string {
	return reflect.TypeOf(dl).String()
}
