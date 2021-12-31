package datatable

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/c-robinson/iplib"
)

type keywordData struct {
	sync.RWMutex
	tag  string
	data []string
}

func newKeywordData(tag string) *keywordData {
	return &keywordData{tag: tag, data: make([]string, 0)}
}

func (k *keywordData) Reset() {
	k.Lock()
	defer k.Unlock()
	k.data = make([]string, 0)
}

func (k *keywordData) ParseFile(r io.Reader) error {
	k.Lock()
	defer k.Unlock()
	k.data = make([]string, 0)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		k.parseline(line)
	}
	return nil
}

func (k *keywordData) ParseLines(lines []string, reset bool) {
	k.Lock()
	defer k.Unlock()
	if reset {
		k.data = make([]string, 0)
	}
	for _, line := range lines {
		k.parseline(line)
	}
}

func (k *keywordData) parseline(line string) {
	if i := strings.IndexByte(line, '#'); i >= 0 {
		line = line[:i]
	}
	attrs := strings.Fields(line)
	if len(attrs) == 1 {
		k.data = append(k.data, line)
		return
	}
	if len(attrs) < 2 {
		return
	}
	if k.tag == strings.ToUpper(attrs[0]) {
		k.data = append(k.data, attrs[1])
	}
}

func (k *keywordData) Match(name string) bool {
	k.RLock()
	defer k.RUnlock()
	for _, k := range k.data {
		if strings.Contains(name, k) {
			return true
		}
	}
	return false
}

func (k *keywordData) LessString() string {
	k.RLock()
	defer k.RUnlock()
	sb := strings.Builder{}
	sb.WriteString("keywordData(Top10):{")
	c := 0
	for _, v := range k.data {
		if c >= 10 {
			break
		} else {
			sb.WriteString(",")
		}
		sb.WriteString(v)
		c += 1
	}
	sb.WriteString("...}")
	return sb.String()
}

func (k *keywordData) ParseInline(ws []string) {
	if k.data == nil {
		k.data = make([]string, 0)
	}
	if len(ws) < 2 {
		fmt.Println("inline len must > 2, format is  tag word...")
		return
	}
	k.Lock()
	defer k.Unlock()
	k.tag = ws[0]
	k.data = ws[1:]
}

func (k *keywordData) Len() int {
	k.RLock()
	defer k.RUnlock()
	return len(k.data)
}

func (k *keywordData) MatchNet(n iplib.Net) bool {
	return false
}

func (k *keywordData) ForEach(f func(interface{}) error, max int) {
	k.RLock()
	defer k.RUnlock()
	c := 0
	for _, datum := range k.data {
		if max > 0 && c >= max {
			break
		}
		_ = f(datum)
	}
}
