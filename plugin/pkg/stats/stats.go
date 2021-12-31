package stats

import (
	"sync"
	"sync/atomic"
)

type Metrics struct {
	Icon  string
	Value interface{}
	Title string
}

func NewMetrics(icon string, value interface{}, title string) *Metrics {
	return &Metrics{Icon: icon, Value: value, Title: title}
}

type Counter struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

func NewCounter(name string, value int64) *Counter {
	return &Counter{Name: name, Value: value}
}

func (d *Counter) Incr(v int64) {
	atomic.AddInt64(&d.Value, v)
}

func (d *Counter) Val() int64 {
	return d.Value
}

type CounterStat struct {
	sync.RWMutex
	cmap map[string]*Counter
}

func NewCounterStat() *CounterStat {
	return &CounterStat{cmap: make(map[string]*Counter)}
}

func (c *CounterStat) Incr(name string, v int64) {
	c.Lock()
	defer c.Unlock()
	if ct, ok := c.cmap[name]; ok {
		ct.Incr(v)
	} else {
		c.cmap[name] = NewCounter(name, v)
	}
}

func (c *CounterStat) MapValues() map[string]int64 {
	mv := make(map[string]int64)
	c.RLock()
	defer c.RUnlock()
	for k, counter := range c.cmap {
		mv[k] = counter.Val()
	}
	return mv
}

func (c *CounterStat) Values() []Counter {
	c.RLock()
	defer c.RUnlock()
	var result = make([]Counter, 0)
	for _, v := range c.cmap {
		result = append(result, *v)
	}
	return result
}

func (c *CounterStat) GetValue(name string) int64 {
	c.RLock()
	defer c.RUnlock()
	if ct, ok := c.cmap[name]; ok {
		return ct.Val()
	} else {
		return 0
	}
}
