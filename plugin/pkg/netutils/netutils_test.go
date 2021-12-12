package netutils

import (
	"testing"

	"github.com/c-robinson/iplib"
)

func Test_compnet(t *testing.T) {
	p1 , _ := ParseIpNet("192.168.0.1/24")
	p2 , _ := ParseIpNet("192.168.0.1/16")
	t.Log(p2.ContainsNet(p1))
}

func Test_CompareNets(t *testing.T) {
	p1 , _ := ParseIpNet("192.168.0.1/20")
	p2 , _ := ParseIpNet("192.168.0.1/24")
	t.Log(iplib.CompareNets(p1, p2))
	t.Log(p1.ContainsNet(p2))
}
