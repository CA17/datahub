package datahub

import (
	"net"
	"testing"
	"time"

	"github.com/ca17/datahub/plugin/pkg/netutils"
	"github.com/miekg/dns"
)

func TestDatahub_MatchGeoip(t *testing.T) {
	dh := NewDatahub()
	dh.geoipPath = "../../data/geoip.dat"
	err := dh.reloadGeoipNetListByTag([]string{"google", "apple", "hk", "private"}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(dh.MatchGeoip("private", net.ParseIP("127.0.0.1")))
	t.Log(dh.MatchGeoip("google", net.ParseIP("8.8.8.8")))
	t.Log(dh.MatchGeoip("google", net.ParseIP("23.200.149.4")))
	t.Log(dh.MatchGeoip("hk", net.ParseIP("23.200.149.4")))
}

func TestDatahub_MatchGeosite(t *testing.T) {
	dh := NewDatahub()
	dh.geositePath = "../../data/geosite.dat"
	err := dh.reloadGeositeDmoainListByTag([]string{"cn"}, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(dh.MatchGeosite(netutils.MatchFullType, "google", "google.com"))
	t.Log(dh.MatchGeosite(netutils.MatchRegexType, "google", "adservice.google.com"))
	t.Log(dh.MatchGeosite(netutils.MatchFullType, "apple", "apple.com"))
	t.Log(dh.MatchGeosite(netutils.MatchFullType, "cn", "qq.com"))
}

func TestDatahub_MatchKeyword(t *testing.T) {
	dh := NewDatahub()
	err := dh.parseKeywordTableByTag("cn", "../../data/keyword_cn.txt")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(dh.MatchKeyword("cn", "www.baidu.com"))
}

func BenchmarkMatchGeositeFull(b *testing.B) {
	dh := NewDatahub()
	dh.geositePath = "../../data/geosite.dat"
	err := dh.reloadGeositeDmoainListByTag([]string{"google", "apple", "hk", "cn"}, true)
	if err != nil {
		b.Fatal(err)
	}
	for n := 0; n < b.N; n++ {
		dh.MatchGeosite(netutils.MatchFullType, "google", "google.cn")
	}
}

func BenchmarkMatchGeositeRegex(b *testing.B) {
	dh := NewDatahub()
	dh.geositePath = "../../data/geosite.dat"
	err := dh.reloadGeositeDmoainListByTag([]string{"google", "apple", "hk"}, true)
	if err != nil {
		b.Fatal(err)
	}
	for n := 0; n < b.N; n++ {
		dh.MatchGeosite(netutils.MatchRegexType, "google", "adservice.google.com")
	}
}

func BenchmarkMatchKeyword(b *testing.B) {
	dh := NewDatahub()
	err := dh.parseKeywordTableByTag("cn", "../../data/keyword_cn.txt")
	if err != nil {
		b.Fatal(err)
	}
	for n := 0; n < b.N; n++ {
		dh.MatchKeyword("cn", "www.baidu.com")
	}
}

func Test_dnslable(t *testing.T) {
	s := "www.google.com"
	i, b := dns.NextLabel(s, 0)
	t.Log(i, b)
	ii, bb := dns.NextLabel(s, i)
	t.Log(ii, bb)
	iii, bbb := dns.NextLabel(s, ii)
	t.Log(iii, bbb)
	t.Logf("%s", string(s[i:iii]))
}

func Test_domainMatch(t *testing.T) {
	fqdn := "qq.www.google.com"
	idx := make([]int, 1, 6)
	off := 0
	end := false

	for {
		off, end = dns.NextLabel(fqdn, off)
		if end {
			break
		}
		idx = append(idx, off)
	}

	for i := range idx {
		p := idx[len(idx)-1-i]
		t.Log(fqdn[p:])
	}
}

func Test_reload(t *testing.T) {
	a := time.Second * 100
	t.Log(a.String())
}
