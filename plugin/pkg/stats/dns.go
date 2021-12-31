package stats

import (
	"sort"
	"sync"
	"time"

	_ "github.com/ahmetb/go-linq"
)

type DayDnsStat struct {
	sync.RWMutex
	LastRolling time.Time
	Interval    time.Duration
	Values      map[int64]map[string]int64
}

func NewDayDnsStat(ival time.Duration) *DayDnsStat {
	return &DayDnsStat{LastRolling: time.Now(), Interval: ival, Values: map[int64]map[string]int64{}}
}

// Rolling 滚动清除老数据
func (d *DayDnsStat) Rolling() {
	d.Lock()
	defer d.Unlock()
	d.LastRolling = time.Now()
	keys := make([]int64, 0)
	for _k, _ := range d.Values {
		keys = append(keys, _k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	s := time.Now().Add(-d.Interval).Unix() * 1000
	for _, key := range keys {
		if s > key {
			d.Values[key] = nil
			delete(d.Values, key)
		}
	}
}

func (d *DayDnsStat) Update(c *CounterStat) {
	if time.Now().Sub(d.LastRolling).Seconds() > 120 {
		d.Rolling()
	}
	v := c.MapValues()
	d.Lock()
	defer d.Unlock()
	if len(v) > 0 {
		d.Values[time.Now().Unix()*1000] = v
	}
}

func (d *DayDnsStat) LineChartData(name string) *LineChartData {
	d.RLock()
	defer d.RUnlock()
	keys := make([]int64, 0)
	for _k, _ := range d.Values {
		keys = append(keys, _k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	ld := NewLineChartData(name, "")
	for _, key := range keys {
		ld.Times = append(ld.Times, time.Unix(key/1000, 0).Format(time.RFC3339))
		for ik, iv := range d.Values[key] {
			if _, ok := ld.Datas[ik]; !ok {
				ld.Datas[ik] = make([]int64, 0)
			}
			ld.Datas[ik] = append(ld.Datas[ik], iv)
		}
	}

	return ld
}
