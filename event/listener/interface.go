package listener

import "github.com/graymonster0927/component/event/event"

type Interface interface {
	Listen() []event.Interface
	Handle(event event.Interface) error
	HandleAsync(event event.Interface)
	IsErrStopped() bool
	Name() string
}
