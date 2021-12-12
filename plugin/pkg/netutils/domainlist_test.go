package netutils

import (
	"fmt"
	"testing"
	"time"
)

func TestDomainList_MatchKey(t *testing.T) {
	var items []string
	for i := 0; i < 10000; i++ {
		items = append(items, fmt.Sprintf("aaa%d", i))
	}
	items = append(items, []string{"abc.com", "sss.net", "xxx.cc"}...)
	for i := 0; i < 10000; i++ {
		items = append(items, fmt.Sprintf("bbb%d", i))
	}
	d3 := NewDomainList()
	d3.InitDomainData(MatchKeywordType, items)
	var start = time.Now()
	if d3.MatchKeyword("abc") {
		t.Log("match key abc")
	}
	t.Log(time.Now().Sub(start).Nanoseconds())

}

func TestDomainList_Match(t *testing.T) {
	d1 := NewDomainList()
	d1.InitDomainData(MatchFullType, []string{"abc.com", "sss.net", "xxx.cc"})
	d1.InitDomainData(MatchRegexType, []string{"(w.)\\.abc\\.com$", "sss.net", "xxx.cc"})
	d1.InitDomainData(MatchKeywordType,  []string{"abc.com", "sss.net", "xxx.cc"})
	if d1.MatchFull("abc") {
		t.Fail()
	}
	if d1.MatchFull("abc.com") {
		t.Log("match full abc.com")
	}
	if d1.MatchRegex("abc") {
		t.Fail()
	}
	if d1.MatchRegex("www.abc.com") {
		t.Log("match reg www.abc.com")
	}
	if d1.MatchKeyword("abc") {
		t.Log("match key abc")
	}
}

func BenchmarkMatchKey(b *testing.B) {
	var items []string
	for i := 0; i < 50000; i++ {
		items = append(items, fmt.Sprintf("aaa%d", i))
	}
	items = append(items, []string{"abc.com", "sss.net", "xxx.cc"}...)
	for i := 0; i < 50000; i++ {
		items = append(items, fmt.Sprintf("bbb%d", i))
	}
	d3 := NewDomainList()
	d3.InitDomainData(MatchKeywordType,items)
	for n := 0; n < b.N; n++ {
		d3.MatchKeyword("abc")
	}
}
