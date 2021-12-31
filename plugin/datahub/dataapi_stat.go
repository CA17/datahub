package datahub

import (
	"github.com/ca17/datahub/plugin/pkg/stats"
)

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
