package loader

import (
	"testing"

	"github.com/ca17/datahub/plugin/pkg/v2data"
)

func TestLoadGeoipCn(t *testing.T) {
	data, err := LoadGeoIPListFromDAT("../../../data/geoip.dat")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, ip := range data.GetEntry() {
		if count >= 200 {
			break
		}
		t.Log(ip.CountryCode)
		// for _, ic := range ip.GetCidr() {
		// 	t.Log(ic.String())
		// }
		// t.Log(ip.String())
		count++
	}
}

func TestLoadGeositeCn(t *testing.T) {
	data, err := LoadGeoSiteList("../../../data/geosite.dat")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, ip := range data.GetEntry() {
		if count >= 200 {
			break
		}
		for _, domain := range ip.GetDomain() {
			if domain.Type == v2data.Domain_Plain {
				t.Log(domain.GetValue())
			}
		}
		count++
	}
}
