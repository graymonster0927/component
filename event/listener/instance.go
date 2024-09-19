package listener

import (
	"sync"
)

var once sync.Once
var instance *EventListener

type EventListener struct {
	listener map[string][]Interface
}

func GetEventListenerInstance() *EventListener {
	once.Do(func() {
		instance = &EventListener{
			listener: make(map[string][]Interface),
		}
	})
	return instance
}

func (el *EventListener) RegisterListener(l Interface) {
	events := l.Listen()
	for _, event := range events {
		if _, ok := el.listener[event.Name()]; ok {
			// todo 去重 listener + 优先级
			el.listener[event.Name()] = append(el.listener[event.Name()], l)
		} else {
			el.listener[event.Name()] = []Interface{l}
		}

	}
}

func (el *EventListener) GetListener(eventName string) []Interface {
	return el.listener[eventName]
}
