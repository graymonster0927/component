package dispatcher

import (
	"errors"
	"fmt"
	"github.com/graymonster0927/component/event/event"
	"github.com/graymonster0927/component/event/listener"
	"sync"
)

var once sync.Once
var instance *Dispatcher

type Dispatcher struct {
	listener *listener.EventListener
}

func GetDispatcherInstance() *Dispatcher {
	once.Do(func() {
		instance = &Dispatcher{
			listener: listener.GetEventListenerInstance(),
		}
	})
	return instance
}

func (d *Dispatcher) Dispatch(event event.Interface) error {
	listeners := d.listener.GetListener(event.Name())
	//同步
	errMsg := ""
	for _, l := range listeners {
		err := l.Handle(event)
		if err != nil {
			errMsg += fmt.Sprintf("err in %s : %s ,", l.Name(), err.Error())
			if l.IsErrStopped() {
				return errors.New(errMsg)
			}
		}
	}

	//异步
	go func() {
		for _, l := range listeners {
			l.HandleAsync(event)
		}
	}()

	if errMsg != "" {
		return errors.New(errMsg)
	}

	return nil
}
