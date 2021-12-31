package datatable

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/allegro/bigcache"
	"github.com/c-robinson/iplib"
	"github.com/metaslink/metasdns/plugin/pkg/netutils"
)

type EcsData struct {
	sync.RWMutex
	tag         string
	netBindings *netutils.NetList
	data        *bigcache.BigCache
}

func NewEcsData(tag string) *EcsData {
	_c, _ := bigcache.NewBigCache(bigcache.DefaultConfig(time.Hour * 24 * 3650))
	return &EcsData{tag: tag, data: _c, netBindings: netutils.NewNetList(make([]iplib.Net, 0))}
}

func (e *EcsData) Reset() {
	e.Lock()
	defer e.Unlock()
	e.data.Reset()
	e.netBindings.Clear()
}

func (e *EcsData) MatchEcsIP(q string) net.IP {
	r, err := e.data.Get(q)
	if err == nil {
		return r
	}
	qnet, err := netutils.ParseIpNet(q)
	if err != nil {
		return nil
	}
	// iter := e.data.Iterator()
	// for iter.SetNext() {
	// 	v, err := iter.Value()
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	fmt.Println(v)
	// }
	if n := e.netBindings.FindNet(qnet); n != nil {
		r, err := e.data.Get(n.String())
		if err == nil {
			return r
		}
	}
	return nil
}

func (e *EcsData) ParseFile(r io.Reader) error {
	e.Lock()
	defer e.Unlock()
	e.data.Reset()
	e.netBindings.Clear()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		// format tag:ip:ipaddr:ecsip || tag:cidr:cidrdata,cidrdata:ecsip
		e.parseLine(line)
	}
	return nil
}

func (e *EcsData) ParseLines(lines []string, reset bool) {
	e.Lock()
	defer e.Unlock()
	if reset {
		e.data.Reset()
		e.netBindings.Clear()
	}
	for _, line := range lines {
		e.parseLine(line)
	}
}

func (e *EcsData) parseLine(line string) {
	if i := strings.IndexByte(line, '#'); i >= 0 {
		line = line[:i]
	}
	if strings.Index(line, ":") != -1 {
		attrs := strings.Split(line, ":")
		if len(attrs) != 3 {
			return
		}
		if e.tag != strings.ToUpper(attrs[0]) {
			return
		}
		var ecsip = net.ParseIP(attrs[2])
		if ecsip == nil {
			return
		}

		var addrs = attrs[1]
		var clist []string
		if strings.Index(addrs, ",") != -1 {
			clist = append(clist, strings.Split(addrs, ",")...)
		} else {
			clist = append(clist, addrs)
		}
		for _, c := range clist {
			if strings.Index(c, "/") != -1 {
				inet, err := netutils.ParseIpNet(c)
				if err != nil {
					continue
				}
				err = e.data.Set(inet.String(), ecsip)
				if err != nil {
					fmt.Println("add ecsip error" + err.Error())
					continue
				}
				e.netBindings.Add(inet)
			} else {
				_ = e.data.Set(c, ecsip)
			}
		}
	}
}

func (e *EcsData) ParseInline(ws []string) {
	fmt.Println("EcsData.ParseInline no support")
}

func (e *EcsData) Match(name string) bool {
	_, err := e.data.Get(name)
	if err != nil {
		return false
	}
	return true
}

func (e *EcsData) MatchNet(inet iplib.Net) bool {
	return e.netBindings.MatchNet(inet)
}

func (e *EcsData) LessString() string {
	sb := strings.Builder{}
	sb.WriteString("EcsData(Top10):{")
	items := make([]string, 0)
	e.netBindings.ForEach(func(inet iplib.Net) {
		items = append(items, inet.String())
	}, 10)
	sb.WriteString(strings.Join(items, ","))
	sb.WriteString("...}")
	return sb.String()
}

func (e *EcsData) Len() int {
	return e.data.Len()
}

func (e *EcsData) ForEach(f func(interface{}) error, max int) {
	e.netBindings.ForEach(func(inet iplib.Net) {
		_ = f(inet)
	}, max)
}
