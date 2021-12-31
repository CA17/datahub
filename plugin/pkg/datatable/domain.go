package datatable

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/c-robinson/iplib"
	"github.com/metaslink/metasdns/plugin/pkg/netutils"
)

type DomainData struct {
	tag  string
	data *netutils.DomainList
}

func newDomainData(tag string) *DomainData {
	return &DomainData{tag: tag, data: netutils.NewDomainList()}
}

func (d *DomainData) Reset() {
	d.data.Clear()
}

func (d *DomainData) ParseFile(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		d.parseline(line)
	}
	return nil
}

func (d *DomainData) ParseLines(lines []string, reset bool) {
	if reset {
		d.data.Clear()
	}
	for _, line := range lines {
		d.parseline(line)
	}
}

func (d *DomainData) parseline(line string) {
	if i := strings.IndexByte(line, '#'); i >= 0 {
		line = line[:i]
	}
	attrs := strings.Fields(line)
	if len(attrs) == 1 {
		d.data.Add(netutils.MatchFullType, line)
		return
	}
	if len(attrs) < 3 {
		return
	}
	if attrs[1] != netutils.MatchFullType &&
		attrs[1] != netutils.MatchDomainType &&
		attrs[1] != netutils.MatchRegexType {
		return
	}
	if d.tag == strings.ToUpper(attrs[0]) {
		d.data.Add(attrs[1], attrs[2])
	}
}

func (d *DomainData) ParseInline(ws []string) {
	if len(ws) < 2 {
		fmt.Println("inline len must > 2, format is  tag word...")
		return
	}
	d.tag = ws[0]
	for _, s := range ws[1:] {
		d.data.Add(netutils.MatchFullType, s)
	}
}

func (d *DomainData) Match(name string) bool {
	return d.data.MixMatch(name)
}

func (d *DomainData) MatchNet(inet iplib.Net) bool {
	return false
}

func (d *DomainData) LessString() string {
	sb := strings.Builder{}
	sb.WriteString("NetlistData(Top10):{")
	items := make([]string, 0)
	d.data.ForEach(func(name string) {
		items = append(items, name)
	}, 10)
	sb.WriteString(strings.Join(items, ","))
	sb.WriteString("...}")
	return sb.String()
}

func (d *DomainData) Len() int {
	return d.data.FullLen() + d.data.RegexLen()
}

func (d *DomainData) ForEach(f func(interface{}) error, max int) {
	d.data.ForEach(func(name string) {
		_ = f(name)
	}, max)
}
