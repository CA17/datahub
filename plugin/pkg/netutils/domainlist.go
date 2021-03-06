package netutils

import (
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/allegro/bigcache"
	"github.com/miekg/dns"
)

const (
	MatchFullType   = "full"
	MatchDomainType = "domain"
	MatchRegexType  = "regex"
)

type DomainList struct {
	sync.RWMutex
	fullTable  *bigcache.BigCache
	regexTable []*regexp.Regexp
}

func NewDomainList() *DomainList {
	_c, _ := bigcache.NewBigCache(bigcache.DefaultConfig(time.Hour * 24 * 3650))
	return &DomainList{
		RWMutex:    sync.RWMutex{},
		fullTable:  _c,
		regexTable: make([]*regexp.Regexp, 0),
	}
}

func (l *DomainList) Add(matchType string, name string) bool {
	switch matchType {
	case MatchFullType, MatchDomainType:
		err := l.fullTable.Set(name, []byte("1"))
		if err != nil {
			return false
		}
		return true
	case MatchRegexType:
		if l.regexTable == nil {
			l.regexTable = make([]*regexp.Regexp, 0)
		}
		re, err := regexp.Compile(name)
		if err == nil {
			l.regexTable = append(l.regexTable, re)
			return true
		}
		return false
	default:
		return false
	}
}

func (l *DomainList) FullLen() int {
	return l.fullTable.Len()
}

func (l *DomainList) RegexLen() int {
	l.RWMutex.RLock()
	defer l.RWMutex.RUnlock()
	return len(l.regexTable)
}

func (l *DomainList) InitDomainData(matchType string, items []string) {
	switch matchType {
	case MatchFullType:
		for _, item := range items {
			_ = l.fullTable.Set(item, []byte("1"))
		}
	case MatchRegexType:
		sort.Strings(items)
		var regexTable []*regexp.Regexp
		for _, item := range items {
			re, err := regexp.Compile(item)
			if err == nil {
				regexTable = append(regexTable, re)
			}
		}
		l.regexTable = regexTable
	default:
		panic("error matchtype")
	}
}

func (l *DomainList) MixMatch(name string) bool {
	if l.MatchFull(name) {
		return true
	}
	if l.MatchDomain(name) {
		return true
	}
	if l.MatchRegex(name) {
		return true
	}
	return false
}

func (l *DomainList) Match(matchType, name string) bool {
	switch matchType {
	case MatchFullType:
		return l.MatchFull(name)
	case MatchDomainType:
		return l.MatchDomain(name)
	case MatchRegexType:
		return l.MatchRegex(name)
	default:
		return false
	}
}

func (l *DomainList) MatchFull(name string) bool {
	_, err := l.fullTable.Get(name)
	if err != nil {
		return false
	}
	return true
}

func (l *DomainList) MatchRegex(name string) bool {
	l.RLock()
	defer l.RUnlock()
	for _, r := range l.regexTable {
		if r.MatchString(name) {
			return true
		}
	}
	return false
}

func (l *DomainList) MatchDomain(name string) bool {
	idx := make([]int, 1, 6)
	off := 0
	end := false
	for {
		off, end = dns.NextLabel(name, off)
		if end {
			break
		}
		idx = append(idx, off)
	}

	for i := range idx {
		p := idx[len(idx)-1-i]
		qname := name[p:]
		r, err := l.fullTable.Get(qname)
		if err == nil && len(r) > 0 {
			return true
		}
	}
	return false
}

func (l *DomainList) ForEach(f func(name string), max int) {
	c := 0
	iter := l.fullTable.Iterator()
	for iter.SetNext() {
		current, err := iter.Value()
		if err == nil {
			if max > 0 && c >= max/2 {
				break
			}
			f(current.Key())
			c++
		}
	}
	l.RLock()
	defer l.RUnlock()
	for _, d := range l.regexTable {
		if max > 0 && c >= max {
			return
		}
		f(d.String())
		c++
	}
}
