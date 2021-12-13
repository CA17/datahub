package netutils

import (
	"testing"
)

func TestDomainList_Match(t *testing.T) {
	d1 := NewDomainList()
	d1.InitDomainData(MatchFullType, []string{"abc.com", "sss.net", "xxx.cc"})
	d1.InitDomainData(MatchRegexType, []string{"(w.)\\.abc\\.com$", "sss.net", "xxx.cc"})
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
}
