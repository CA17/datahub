package datatable

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/c-robinson/iplib"
	"github.com/ca17/datahub/plugin/pkg/netutils"
)

type DomainData struct {
	tag  string
	data *netutils.DomainList
}

func newDomainData(tag string) *DomainData {
	return &DomainData{tag: tag}
}

func (d *DomainData) ParseFile(r io.Reader) error {
	d.data = netutils.NewDomainList()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}

		if strings.Index(line, ":") != -1 {
			attrs := strings.Split(line, ":")
			if len(attrs) != 3 {
				continue
			}
			if attrs[1] != netutils.MatchFullType &&
				attrs[1] != netutils.MatchDomainType &&
				attrs[1] != netutils.MatchRegexType {
				continue
			}
			if d.tag == strings.ToUpper(attrs[0]) {
				d.data.Add(attrs[1], attrs[2])
			}
		} else {
			d.data.Add(netutils.MatchFullType, line)
		}
	}
	return nil
}

func (d *DomainData) ParseLines(lines []string) {
	if d.data == nil {
		d.data = netutils.NewDomainList()
	}
	for _, line := range lines {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		if strings.Index(line, ":") != -1 {
			attrs := strings.Split(line, ":")
			if len(attrs) != 3 {
				continue
			}
			if attrs[1] != netutils.MatchFullType &&
				attrs[1] != netutils.MatchDomainType &&
				attrs[1] != netutils.MatchRegexType {
				continue
			}
			if d.tag == strings.ToUpper(attrs[0]) {
				d.data.Add(attrs[1], attrs[2])
			}
		} else {
			d.data.Add(netutils.MatchFullType, line)
		}
	}
}

func (d *DomainData) ParseInline(ws []string) {
	if d.data == nil {
		d.data = netutils.NewDomainList()
	}
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
