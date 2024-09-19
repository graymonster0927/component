package event

import "reflect"

type DemoAsyncEvent struct {
}

func (de *DemoAsyncEvent) Name() string {
	return reflect.TypeOf(de).String()
}
