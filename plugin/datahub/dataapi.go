package datahub

import (
	"net"
	"strings"
	"time"

	"github.com/c-robinson/iplib"
	"github.com/ca17/datahub/plugin/pkg/datatable"
	"github.com/ca17/datahub/plugin/pkg/loader"
	"github.com/ca17/datahub/plugin/pkg/netutils"
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
	inet := iplib.NewNet(ip, 32)
	if list := dh.getGeoNetListByTag(tag); list != nil {
		return list.MatchNet(inet)
	}
	return false
}

func (dh *Datahub) MatchGeoNet(tag string, net iplib.Net) bool {
	tag = strings.ToUpper(tag)
	if list := dh.getGeoNetListByTag(tag); list != nil {
		return list.MatchNet(net)
	}
	return false
}

// MixMatchNetByStr 混合模式匹配网络地址
func (dh *Datahub) MixMatchNetByStr(tag string, ns string) bool {
	inet, err := netutils.ParseIpNet(ns)
	if err != nil {
		return false
	}
	return dh.MixMatchNet(tag, inet)
}

// MixMatchNet 混合模式匹配网络地址
func (dh *Datahub) MixMatchNet(tag string, ns iplib.Net) bool {
	tag = strings.ToUpper(tag)
	// 匹配自定义网络地址列表
	if list := dh.getDataTableByTag(datatable.DateTypeNetlistTable, tag); list != nil &&
		list.GetData().(*datatable.NetlistData).MatchNet(ns) {
		return true
	}

	// 匹配Geodat网络地址列表
	if list := dh.getGeoNetListByTag(tag); list != nil && list.MatchNet(ns) {
		return true
	}

	return false
}

// MatchNetByStr 匹配自定义网络地址
func (dh *Datahub) MatchNetByStr(tag string, ns string) bool {
	inet, err := netutils.ParseIpNet(ns)
	if err != nil {
		return false
	}
	return dh.MatchNet(tag, inet)
}

// MatchNet 匹配自定义网络地址
func (dh *Datahub) MatchNet(tag string, ns iplib.Net) bool {
	tag = strings.ToUpper(tag)
	// 匹配自定义网络地址列表
	if list := dh.getDataTableByTag(datatable.DateTypeNetlistTable, tag); list != nil &&
		list.GetData().(*datatable.NetlistData).MatchNet(ns) {
		return true
	}
	return false
}

// MatchGeosite 匹配 Geosite 域名
func (dh *Datahub) MatchGeosite(matchType, tag string, name string) bool {
	tag = strings.ToUpper(tag)
	if list := dh.getGeoDomainListByTag(tag); list != nil {
		return list.Match(matchType, name)
	}
	return false
}

// MatchKeyword 域名关键词匹配
func (dh *Datahub) MatchKeyword(tag string, name string) bool {
	tag = strings.ToUpper(tag)
	if list := dh.getDataTableByTag(datatable.DateTypeKeywordTable, tag); list != nil {
		return list.Match(name)
	}
	return false
}

// MatchEcs 匹配 ECS IP
func (dh *Datahub) MatchEcs(tag string, client string) net.IP {
	tag = strings.ToUpper(tag)
	if list := dh.getDataTableByTag(datatable.DateTypeEcsTable, tag); list != nil {
		return list.GetData().(*datatable.EcsData).MatchEcsIP(client)
	}
	return nil
}

// MixMatch 混合模式匹配域名
func (dh *Datahub) MixMatch(tag string, name string) bool {
	tag = strings.ToUpper(tag)

	// 匹配自定义域名表
	if list := dh.getDataTableByTag(datatable.DateTypeDomainlistTable, tag); list != nil {
		if list.Match(name) {
			_ = dh.matchCache.Set(tag+name, _mateched)
			return true
		}
	}

	// 匹配关键词表
	if list := dh.getDataTableByTag(datatable.DateTypeKeywordTable, tag); list != nil {
		if list.Match(name) {
			return true
		}
	}

	// 匹配 Geodat 数据
	if list := dh.getGeoDomainListByTag(tag); list != nil {
		if list.MixMatch(name) {
			return true
		}
	}

	return false
}

var _mateched = []byte("1")

// MixMatchTags 混合模式匹配域名
func (dh *Datahub) MixMatchTags(tags []string, name string) bool {
	if len(tags) == 0 {
		return false
	}
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

		// 匹配自定义域名表
		if list := dh.getDataTableByTag(datatable.DateTypeDomainlistTable, tag); list != nil {
			if list.Match(name) {
				_ = dh.matchCache.Set(tag+name, _mateched)
				return true
			}
		}

		// 匹配关键词表
		if list := dh.getDataTableByTag(datatable.DateTypeKeywordTable, tag); list != nil {
			if list.Match(name) {
				_ = dh.matchCache.Set(tag+name, _mateched)
				return true
			}
		}

		// 匹配 Geodat 数据
		if list := dh.getGeoDomainListByTag(tag); list != nil {
			if list.MixMatch(name) {
				_ = dh.matchCache.Set(tag+name, _mateched)
				return true
			}
		}

	}
	return false
}
