package datahub

import (
	"net"
	"strings"
	"time"

	"github.com/c-robinson/iplib"
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
	dh.nlmLock.RLock()
	defer dh.nlmLock.RUnlock()
	inet := iplib.NewNet(ip, 32)
	if list, ok := dh.geoipNetListMap[tag]; ok {
		return list.MatchNet(inet)
	}
	return false
}

func (dh *Datahub) MatchGeoNet(tag string, net iplib.Net) bool {
	tag = strings.ToUpper(tag)
	dh.nlmLock.RLock()
	defer dh.nlmLock.RUnlock()
	if list, ok := dh.geoipNetListMap[tag]; ok {
		return list.MatchNet(net)
	}
	return false
}

// MatchGeosite 匹配 Geosite 域名
func (dh *Datahub) MatchGeosite(matchType, tag string, name string) bool {
	tag = strings.ToUpper(tag)
	dh.dlmLock.RLock()
	defer dh.dlmLock.RUnlock()
	if list, ok := dh.geositeDoaminListMap[tag]; ok {
		return list.Match(matchType, name)
	}
	return false
}

func (dh *Datahub) MatchKeyword(tag string, name string) bool {
	tag = strings.ToUpper(tag)
	dh.ktLock.RLock()
	defer dh.ktLock.RUnlock()
	if list := dh.getKeywordTableByTag(tag); list != nil {
		return list.Match(name)
	}
	return false
}

// MixMatch 混合模式匹配域名
func (dh *Datahub) MixMatch(tag string, name string) bool {
	tag = strings.ToUpper(tag)

	if list := dh.getDomainListByTag(tag); list != nil {
		if list.MixMatch(name) {
			return true
		}
	}

	if list := dh.getKeywordTableByTag(tag); list != nil {
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

		if list := dh.getDomainListByTag(tag); list != nil {
			if list.MixMatch(name) {
				_ = dh.matchCache.Set(tag+name, _mateched)
				return true
			}
		}
		if list := dh.getKeywordTableByTag(tag); list != nil {
			if list.Match(name) {
				_ = dh.matchCache.Set(tag+name, _mateched)
				return true
			}
		}
	}
	return false
}
