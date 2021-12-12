package loader

import (
	"strings"
	"testing"
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
		if ip.GetCountryCode() != "CN" {
			continue
		}
		for _, domain := range ip.GetDomain() {
			if strings.Contains(domain.GetValue(), "qq.com") {
				t.Log(domain.GetValue())
			}
		}
		count++
	}
}
