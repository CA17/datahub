package datatable

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/c-robinson/iplib"
	"github.com/metaslink/metasdns/plugin/pkg/netutils"
)

type NetlistData struct {
	tag  string
	data *netutils.NetList
}

func newNetlistData(tag string) *NetlistData {
	return &NetlistData{tag: tag, data: &netutils.NetList{}}
}

func (n *NetlistData) Reset() {
	n.data.Clear()
}

func (n *NetlistData) ParseFile(r io.Reader) error {
	n.data.Clear()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		n.parseline(line)
	}
	return nil
}

func (n *NetlistData) ParseLines(lines []string, reset bool) {
	if reset {
		n.data.Clear()
	}
	for _, line := range lines {
		n.parseline(line)
	}
}

func (n *NetlistData) parseline(line string) {
	if i := strings.IndexByte(line, '#'); i >= 0 {
		line = line[:i]
	}
	attrs := strings.Fields(line)
	if len(attrs) == 1 {
		n.data.AddByString(line)
		return
	}
	if len(attrs) < 2 {
		return
	}
	if n.tag == strings.ToUpper(attrs[0]) {
		n.data.AddByString(attrs[1])
	}
}

func (n *NetlistData) ParseInline(ws []string) {
	if len(ws) < 2 {
		fmt.Println("inline len must > 2, format is  tag word...")
		return
	}
	n.tag = ws[0]
	for _, s := range ws[1:] {
		n.data.AddByString(s)
	}
}

func (n *NetlistData) Match(name string) bool {
	inet, err := netutils.ParseIpNet(name)
	if err != nil {
		return false
	}
	return n.MatchNet(inet)
}

func (n *NetlistData) MatchNet(inet iplib.Net) bool {
	return n.data.MatchNet(inet)
}

func (n *NetlistData) LessString() string {
	sb := strings.Builder{}
	sb.WriteString("NetlistData(Top10):{")
	items := make([]string, 0)
	n.data.ForEach(func(inet iplib.Net) {
		items = append(items, inet.String())
	}, 10)
	sb.WriteString(strings.Join(items, ","))
	sb.WriteString("...}")
	return sb.String()
}

func (n *NetlistData) Len() int {
	return n.data.Len()
}

func (n *NetlistData) ForEach(f func(interface{}) error, max int) {
	n.data.ForEach(func(item iplib.Net) {
		_ = f(item)
	}, max)
}
