package main

import (
	_ "github.com/ca17/dnssrc/plugin/dnssrc"
	"github.com/coredns/coredns/core/dnsserver"
	_ "github.com/coredns/coredns/core/plugin"
	"github.com/coredns/coredns/coremain"
)

func index(slice []string, item string) int {
    for i := range slice {
        if slice[i] == item {
            return i
        }
    }
    return -1
}

func main() {
	// insert dnssrc before forward
	idx := index(dnsserver.Directives, "geoip")
	dnsserver.Directives = append(dnsserver.Directives[:idx], append([]string{"datahub"}, dnsserver.Directives[idx:]...)...)
	coremain.Run()
}
