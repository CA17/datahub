package datahub

import (
	"testing"

	"github.com/coredns/caddy"
)

func Test_parseConfig(t *testing.T) {
	c := caddy.NewTestController("dns", `datahub {
        debug
        geoip_path ../../data/geoip.dat
        geosite_path ../../data/geosite.dat
        geoip_cache cn hk jp google apple
        geosite_cache cn hk jp private apple
        geodat_upgrade_url http://xxxx.com
        geodat_upgrade_cron 0 30 0 * * *
        keyword_table cn ../../data/keyword_cn.txt
        domain_table cn ../../data/domain_cn.txt
        netlist_table cn ../../data/netlist_cn.txt
        ecs_table global ../../data/ecs_table.txt
        reload @every 3s
    }`)
	ecs, err := parseConfig(c)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ecs)
}
