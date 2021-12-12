package datatable

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
)

type keywordData struct {
	sync.RWMutex
	tag  string
	data []string
}

func newKeywordData(tag string) *keywordData {
	return &keywordData{tag: tag}
}

func (k *keywordData) ParseFile(r io.Reader) error {
	k.Lock()
	defer k.Unlock()
	k.data = make([]string, 0)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}

		if strings.Index(line, ":") != -1 {
			attrs := strings.Split(line, ":")
			if len(attrs) != 2 {
				continue
			}
			if k.tag == strings.ToUpper(attrs[0]){
				k.data = append(k.data, attrs[1])
			}
		} else {
			k.data = append(k.data, line)
		}
	}
	return nil
}

func (k *keywordData) ParseLines(lines []string) {
	k.Lock()
	defer k.Unlock()
	for _, line := range lines {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		if strings.Index(line, ":") != -1 {
			attrs := strings.Split(line, ":")
			if len(attrs) != 2 {
				continue
			}
			if k.tag == strings.ToUpper(attrs[0]) {
				k.data = append(k.data, attrs[1])
			}
		} else {
			k.data = append(k.data, line)
		}
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
	sb.WriteString("keywordData:{")
	c := 0
	for _, v := range k.data {
		if c >= 10 {
			sb.WriteString("......")
			break
		}
		sb.WriteString(v)
		sb.WriteString(",")
		c += 1
	}
	return sb.String()
}

func (k *keywordData) ParseInline(ws []string) {
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
