package netutils

import (
	"sort"
	"sync"

	"github.com/c-robinson/iplib"
)

type NetList struct {
	sync.RWMutex
	data []iplib.Net
}

func NewNetList(data []iplib.Net) *NetList {
	return &NetList{data: data, RWMutex: sync.RWMutex{}}
}

func (l *NetList) Len() int {
	l.RLock()
	defer l.RUnlock()
	return len(l.data)
}

func (l *NetList) Add(inet iplib.Net) {
	l.Lock()
	defer l.Unlock()
	l.data = append(l.data, inet)
}

func (l *NetList) Sort() {
	l.Lock()
	defer l.Unlock()
	sort.Slice(l.data, func(i, j int) bool {
		return iplib.CompareNets(l.data[i], l.data[j]) < 1
	})
}

func (l *NetList) MatchNet(lookingFor iplib.Net) bool {
	l.RLock()
	defer l.RUnlock()
	var low int = 0
	var high int = len(l.data) - 1
	for low <= high {
		var mid int = low + (high-low)/2
		var midValue = l.data[mid]
		if midValue.ContainsNet(lookingFor) {
			return mid != -1
		} else if iplib.CompareNets(midValue, lookingFor) > 0 {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return false
}
