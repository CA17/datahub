package datahub

import (
	"net"
	"strings"
	"time"

	"github.com/c-robinson/iplib"
	"github.com/ca17/datahub/plugin/pkg/datatable"
	"github.com/ca17/datahub/plugin/pkg/loader"
	"github.com/ca17/datahub/plugin/pkg/v2data"
)

// LoadGeoIPFromDAT 载入 GEOIP 数据表
func (dh *Datahub) LoadGeoIPFromDAT(tag string) (*v2data.GeoIP, error) {
	return loader.LoadGeoIPFromDAT(dh.geoipPath, tag)
}

// LoadGeoSiteFromDAT 载入 GEOSITE 表
func (dh *Datahub) LoadGeoSiteFromDAT(country string) (*v2data.GeoSite, error) {
	return loader.LoadGeoSiteFromDAT(dh.geositePath, country)
}

func (dh *Datahub) MatchGeoip(tag string, ip net.IP) bool {
	tag = strings.ToUpper(tag)
	dh.geonlmLock.RLock()
	defer dh.geonlmLock.RUnlock()
	inet := iplib.NewNet(ip, 32)
	if list, ok := dh.geoipNetListMap[tag]; ok {
		return list.MatchNet(inet)
	}
	return false
}

func (dh *Datahub) MatchGeoNet(tag string, net iplib.Net) bool {
	tag = strings.ToUpper(tag)
	dh.geonlmLock.RLock()
	defer dh.geonlmLock.RUnlock()
	if list, ok := dh.geoipNetListMap[tag]; ok {
		return list.MatchNet(net)
	}
	return false
}

// MatchGeosite 匹配 Geosite 域名
func (dh *Datahub) MatchGeosite(matchType, tag string, name string) bool {
	tag = strings.ToUpper(tag)
	dh.geodlmLock.RLock()
	defer dh.geodlmLock.RUnlock()
	if list, ok := dh.geositeDoaminListMap[tag]; ok {
		return list.Match(matchType, name)
	}
	return false
}

func (dh *Datahub) MatchKeyword(tag string, name string) bool {
	tag = strings.ToUpper(tag)
	if list := dh.getDataTableByTag(datatable.DateTypeKeywordTable, tag); list != nil {
		return list.Match(name)
	}
	return false
}

func (dh *Datahub) MatchEcs(tag string, name string) net.IP {
	tag = strings.ToUpper(tag)
	if list := dh.getDataTableByTag(datatable.DateTypeEcsTable, tag); list != nil {
		return list.GetData().(*datatable.EcsData).MatchEcsIP(name)
	}
	return nil
}

// MixMatch 混合模式匹配域名
func (dh *Datahub) MixMatch(tag string, name string) bool {
	tag = strings.ToUpper(tag)

	if list := dh.getGeoDomainListByTag(tag); list != nil {
		if list.MixMatch(name) {
			return true
		}
	}

	if list := dh.getDataTableByTag(datatable.DateTypeKeywordTable, tag); list != nil {
		if list.Match(name) {
			return true
		}
	}

	return false
}

var _mateched = []byte("1")

// MixMatchTags 混合模式匹配域名
func (dh *Datahub) MixMatchTags(tags []string, name string) bool {
	var start = time.Now()
	defer func() {
		log.Debugf("Match %s cast %d ns", name, time.Now().Sub(start).Nanoseconds())
	}()
	for _, tag := range tags {
		tag = strings.ToUpper(tag)

		_, err := dh.matchCache.Get(tag + name)
		if err == nil {
			return true
		}

		if list := dh.getGeoDomainListByTag(tag); list != nil {
			if list.MixMatch(name) {
				_ = dh.matchCache.Set(tag+name, _mateched)
				return true
			}
		}
		if list := dh.getDataTableByTag(datatable.DateTypeKeywordTable, tag); list != nil {
			if list.Match(name) {
				_ = dh.matchCache.Set(tag+name, _mateched)
				return true
			}
		}
	}
	return false
}
