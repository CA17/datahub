package netutils

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/c-robinson/iplib"
)

var testdata NetList

func initData() *NetList {
	var cc []iplib.Net
	for j := 0; j <= 200; j++ {
		for c := 0; c <= 250; c++ {
			cc = append(cc, iplib.Net4FromStr(fmt.Sprintf("%d.%d.0.0/24", c, j)))
			cc = append(cc, iplib.Net4FromStr(fmt.Sprintf("1.%d.0.0/24", c)))
			cc = append(cc, iplib.Net4FromStr(fmt.Sprintf("10.%d.0.0/24", c)))
			cc = append(cc, iplib.Net4FromStr(fmt.Sprintf("10.0.%d.0/24", c)))
			cc = append(cc, iplib.Net4FromStr(fmt.Sprintf("100.%d.%d.0/24", j, c)))
			cc = append(cc, iplib.Net4FromStr(fmt.Sprintf("12.%d.%d.0/24", j, c)))
			cc = append(cc, iplib.Net4FromStr(fmt.Sprintf("13.%d.%d.0/24", j, c)))
			cc = append(cc, iplib.Net4FromStr(fmt.Sprintf("14.%d.%d.0/24", j, c)))
			c++
		}
	}
	sort.Slice(cc, func(i, j int) bool {
		return iplib.CompareNets(cc[i], cc[j]) < 1
	})
	fmt.Println(len(cc))
	data := NewNetList(cc)
	data.Sort()
	// fmt.Println(data.Data)
	return data
}

func TestCompNet(t *testing.T) {

}

func TestNetList_MatchNetat(t *testing.T) {
	ns := initData()
	var start = time.Now()
	s, _ := ParseIpNet("10.0.200.2")
	i := ns.MatchNet(s)
	t.Log(time.Now().Sub(start).Nanoseconds())
	t.Log(i)
}

func BenchmarkMatchNet(b *testing.B) {
	ns := initData()
	s, _ := ParseIpNet("10.0.200.2")
	for n := 0; n < b.N; n++ {
		ns.MatchNet(s)
	}
}

