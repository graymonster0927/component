package safe_counter

import "sync"

type SafeCounter struct {
	value map[string]uint64 // 计数器的值
	mux   sync.Mutex        // 互斥锁，保护计数器的线程安全
}

// Inc 方法用于增加键的计数
func (c *SafeCounter) Inc(key string) {
	c.mux.Lock()   // 加锁
	c.value[key]++ // 增加计数
	c.mux.Unlock() // 解锁
}

// Inc 方法用于增加键的计数
func (c *SafeCounter) IncN(key string, n uint64) {
	c.mux.Lock()                    // 加锁
	c.value[key] = c.value[key] + n // 增加计数
	c.mux.Unlock()                  // 解锁
}

// Value 方法返回某个键的计数
func (c *SafeCounter) Value(key string) uint64 {
	c.mux.Lock()         // 加锁
	defer c.mux.Unlock() // 延迟解锁
	return c.value[key]  // 返回计数
}

func (c *SafeCounter) All(withClear bool) map[string]uint64 {
	c.mux.Lock()         // 加锁
	defer c.mux.Unlock() // 延迟解锁
	tmp := c.value
	if withClear {
		c.value = make(map[string]uint64)
	}
	return tmp
}

func NewSafeCounter() *SafeCounter {
	return &SafeCounter{
		value: make(map[string]uint64),
		mux:   sync.Mutex{},
	}
}
