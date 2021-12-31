package datahub

import (
	"net"
	"strings"
	"time"

	"github.com/c-robinson/iplib"
	"github.com/ca17/datahub/plugin/pkg/datatable"
	"github.com/ca17/datahub/plugin/pkg/loader"
	"github.com/ca17/datahub/plugin/pkg/netutils"
	"github.com/ca17/datahub/plugin/pkg/stats"
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
		if list.MatchNet(inet) {
			dh.networkMatchStat.Incr(tag, 1)
			return true
		}
	}
	return false
}

func (dh *Datahub) MatchGeoNet(tag string, net iplib.Net) bool {
	tag = strings.ToUpper(tag)
	if list := dh.getGeoNetListByTag(tag); list != nil {
		if list.MatchNet(net) {
			dh.networkMatchStat.Incr(tag, 1)
			return true
		}
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
		dh.networkMatchStat.Incr(tag, 1)
		return true
	}

	// 匹配Geodat网络地址列表
	if list := dh.getGeoNetListByTag(tag); list != nil && list.MatchNet(ns) {
		dh.networkMatchStat.Incr(tag, 1)
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
		dh.networkMatchStat.Incr(tag, 1)
		return true
	}
	return false
}

// MatchGeosite 匹配 Geosite 域名
func (dh *Datahub) MatchGeosite(matchType, tag string, name string) bool {
	tag = strings.ToUpper(tag)
	if list := dh.getGeoDomainListByTag(tag); list != nil {
		if list.Match(matchType, name) {
			dh.domainMatchStat.Incr(tag, 1)
			return true
		}
	}
	return false
}

// MatchKeyword 域名关键词匹配
func (dh *Datahub) MatchKeyword(tag string, name string) bool {
	tag = strings.ToUpper(tag)
	if list := dh.getDataTableByTag(datatable.DateTypeKeywordTable, tag); list != nil {
		if list.Match(name) {
			dh.keywordMatchStat.Incr(tag, 1)
			return true
		}
	}
	return false
}

// MatchEcs 匹配 ECS IP
func (dh *Datahub) MatchEcs(tag string, client string) net.IP {
	tag = strings.ToUpper(tag)
	if list := dh.getDataTableByTag(datatable.DateTypeEcsTable, tag); list != nil {
		m := list.GetData().(*datatable.EcsData)
		ecsip := m.MatchEcsIP(client)
		if ecsip != nil {
			dh.metricsStat.Incr(MetricsStatEcsHits, 1)
			return ecsip
		}
	}
	return nil
}

// MixMatch 混合模式匹配域名
func (dh *Datahub) MixMatch(tag string, name string) bool {
	tag = strings.ToUpper(tag)

	bs, err := dh.matchCache.Get(tag + name)
	if err == nil {
		switch string(bs) {
		case DomainMatcher:
			dh.domainMatchStat.Incr(tag, 1)
		case KeywordMatcher:
			dh.keywordMatchStat.Incr(tag, 1)
		}
		return true
	}

	// 匹配自定义域名表
	if list := dh.getDataTableByTag(datatable.DateTypeDomainlistTable, tag); list != nil {
		if list.Match(name) {
			_ = dh.matchCache.Set(tag+name, []byte(DomainMatcher))
			dh.domainMatchStat.Incr(tag, 1)
			return true
		}
	}

	// 匹配关键词表
	if list := dh.getDataTableByTag(datatable.DateTypeKeywordTable, tag); list != nil {
		if list.Match(name) {
			_ = dh.matchCache.Set(tag+name, []byte(KeywordMatcher))
			dh.keywordMatchStat.Incr(tag, 1)
			return true
		}
	}

	// 匹配 Geodat 数据
	if list := dh.getGeoDomainListByTag(tag); list != nil {
		if list.MixMatch(name) {
			_ = dh.matchCache.Set(tag+name, []byte(DomainMatcher))
			dh.domainMatchStat.Incr(tag, 1)
			return true
		}
	}

	return false
}

var reverseResult = func(rval, reverse bool) bool {
	if reverse {
		return !rval
	}
	return rval
}

// MixMatchTags 混合模式匹配域名
func (dh *Datahub) MixMatchTags(tags []string, name string, reverse bool) bool {
	if len(tags) == 0 {
		return false
	}
	var start = time.Now()
	defer func() {
		log.Debugf("Match %s cast %d ns", name, time.Now().Sub(start).Nanoseconds())
	}()
	for _, tag := range tags {
		tag = strings.ToUpper(tag)
		bs, err := dh.matchCache.Get(tag + name)
		if err == nil {
			switch string(bs) {
			case DomainMatcher:
				dh.domainMatchStat.Incr(tag, 1)
			case KeywordMatcher:
				dh.keywordMatchStat.Incr(tag, 1)
			}
			return reverseResult(true, reverse)
		}
		// 匹配自定义域名表
		if list := dh.getDataTableByTag(datatable.DateTypeDomainlistTable, tag); list != nil {
			if list.Match(name) {
				_ = dh.matchCache.Set(tag+name, []byte(DomainMatcher))
				dh.domainMatchStat.Incr(tag, 1)
				return reverseResult(true, reverse)
			}
		}
		// 匹配关键词表
		if list := dh.getDataTableByTag(datatable.DateTypeKeywordTable, tag); list != nil {
			if list.Match(name) {
				_ = dh.matchCache.Set(tag+name, []byte(KeywordMatcher))
				dh.keywordMatchStat.Incr(tag, 1)
				return reverseResult(true, reverse)
			}
		}
		// 匹配 Geodat 数据
		if list := dh.getGeoDomainListByTag(tag); list != nil {
			if list.MixMatch(name) {
				_ = dh.matchCache.Set(tag+name, []byte(DomainMatcher))
				dh.domainMatchStat.Incr(tag, 1)
				return reverseResult(true, reverse)
			}
		}
		if reverse {
			dh.domainMatchStat.Incr("!"+tag, 1)
		}
	}
	return reverseResult(false, reverse)
}

// stats

// IncrMetricsCounter metadnsq 调用，递增Metrics计数器
func (dh *Datahub) IncrMetricsCounter(name string) {
	dh.metricsStat.Incr(name, 1)
}

// GetMetricSValue 查询 Metrics 统计值
func (dh *Datahub) GetMetricSValue(name string) int64 {
	return dh.metricsStat.GetValue(name)
}

// GetDomainDayLineStat 查询域名标签匹配 24小时趋势图
func (dh *Datahub) GetDomainDayLineStat() *stats.LineChartData {
	return dh.dayaDomainChartStat.LineChartData("最近 24 小时域名匹配统计").ChartData()
}

// GetNetworkDayLineStat 查询网络标签匹配 24小时趋势图
func (dh *Datahub) GetNetworkDayLineStat() *stats.LineChartData {
	return dh.dayaDomainChartStat.LineChartData("最近 24 小时网络地址匹配统计").ChartData()
}

// QueryTotal 查询DNS 请求总数统计值
func (dh *Datahub) QueryTotal() int64 {
	return dh.metricsStat.GetValue(MetricsStatDnsQuery)
}

// MatcherStats 查询 Metrics 指标统计
func (dh *Datahub) MatcherStats(classify string) []stats.Counter {
	switch classify {
	case "domain":
		return dh.domainMatchStat.Values()
	case "keyword":
		return dh.keywordMatchStat.Values()
	case "network":
		return dh.networkMatchStat.Values()
	default:
		return []stats.Counter{
			*stats.NewCounter("unknow", 0),
		}
	}
}
