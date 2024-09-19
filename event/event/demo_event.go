package event

import "reflect"

type DemoEvent struct {
}

func (de *DemoEvent) Name() string {
	return reflect.TypeOf(de).String()
}
