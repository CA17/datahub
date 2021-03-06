package datahub

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/allegro/bigcache"
	"github.com/c-robinson/iplib"
	"github.com/ca17/datahub/plugin/pkg/datatable"
	"github.com/ca17/datahub/plugin/pkg/loader"
	"github.com/ca17/datahub/plugin/pkg/netutils"
	"github.com/ca17/datahub/plugin/pkg/stats"
	"github.com/ca17/datahub/plugin/pkg/v2data"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"github.com/orcaman/concurrent-map"
	"github.com/robfig/cron/v3"
)

const (
	DomainMatcher  = "domain"
	NetworkMatcher = "network"
	KeywordMatcher = "keyword"

	MetricsStatDnsQuery = "dnsquery"
	MetricsStatEcsHits  = "ecshits"
	MetricsStatNxdomain = "nxdomain"
)

var cronParser = cron.NewParser(
	cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

type Datahub struct {
	bootstrap            []string
	geonlmLock           sync.RWMutex
	geodlmLock           sync.RWMutex
	geoipNetListMap      map[string]*netutils.NetList
	geositeDoaminListMap map[string]*netutils.DomainList
	keywordTableMap      cmap.ConcurrentMap
	netlistTableMap      cmap.ConcurrentMap
	domainTableMap       cmap.ConcurrentMap
	ecsTableMap          cmap.ConcurrentMap
	//
	Next              plugin.Handler
	geoipCacheTags    []string
	geositeCacheTags  []string
	geoipPath         string
	geositePath       string
	geodatUpgradeUrl  string
	geodatUpgradeCron string
	sched             *cron.Cron
	matchCache        *bigcache.BigCache
	reloadCron        string
	pubserver         *dataServer
	notifyServer      *notifyServer
	jwtSecret         string
	debug             bool

	// stat define
	metricsStat         *stats.CounterStat // metrics 统计
	queryStat           *stats.CounterStat // 查询域名统计
	clientStat          *stats.CounterStat // 客户端统计
	domainMatchStat     *stats.CounterStat // 域名标签统计
	networkMatchStat    *stats.CounterStat // 网络标签统计
	keywordMatchStat    *stats.CounterStat // 关键词统计
	dayaDomainChartStat *stats.DayDnsStat  // 最近24小时域名标签匹配数统计
	dayNetworkChartStat *stats.DayDnsStat  // 最近24小时网络标签匹配数统计
}

// ServeDNS Datahub 只做简单匹配统计， 不会处理请求逻辑
func (dh *Datahub) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := &request.Request{W: w, Req: r}
	dh.metricsStat.Incr(MetricsStatDnsQuery, 1)
	dh.queryStat.Incr(state.Name(), 1)
	dh.clientStat.Incr(state.IP(), 1)
	return plugin.NextOrFailure(dh.Name(), dh.Next, ctx, w, r)
}

func (dh *Datahub) Name() string { return "datahub" }

func NewDatahub() *Datahub {
	mc, _ := bigcache.NewBigCache(bigcache.DefaultConfig(time.Second * 300))
	hub := &Datahub{
		geonlmLock:           sync.RWMutex{},
		geodlmLock:           sync.RWMutex{},
		geoipNetListMap:      make(map[string]*netutils.NetList),
		geositeDoaminListMap: make(map[string]*netutils.DomainList),
		keywordTableMap:      cmap.New(),
		netlistTableMap:      cmap.New(),
		domainTableMap:       cmap.New(),
		ecsTableMap:          cmap.New(),
		matchCache:           mc,
		notifyServer:         newNotifyServer(),
		sched:                cron.New(cron.WithParser(cronParser)),
		// stat
		domainMatchStat:     stats.NewCounterStat(),
		networkMatchStat:    stats.NewCounterStat(),
		keywordMatchStat:    stats.NewCounterStat(),
		metricsStat:         stats.NewCounterStat(),
		queryStat:           stats.NewCounterStat(),
		clientStat:          stats.NewCounterStat(),
		dayaDomainChartStat: stats.NewDayDnsStat(time.Hour * 24),
		dayNetworkChartStat: stats.NewDayDnsStat(time.Hour * 24),
	}
	hub.pubserver = newPubServer(hub)
	return hub
}

func (dh *Datahub) getGeoNetListByTag(tag string) *netutils.NetList {
	dh.geodlmLock.RLock()
	defer dh.geodlmLock.RUnlock()
	return dh.geoipNetListMap[tag]
}

func (dh *Datahub) getGeoDomainListByTag(tag string) *netutils.DomainList {
	dh.geodlmLock.RLock()
	defer dh.geodlmLock.RUnlock()
	return dh.geositeDoaminListMap[tag]
}

func (dh *Datahub) getDataTableByTag(dtyope string, tag string) *datatable.DataTable {
	tag = strings.ToUpper(tag)
	switch dtyope {
	case datatable.DateTypeEcsTable:
		if v, ok := dh.ecsTableMap.Get(tag); ok {
			return v.(*datatable.DataTable)
		}
	case datatable.DateTypeNetlistTable:
		if v, ok := dh.netlistTableMap.Get(tag); ok {
			return v.(*datatable.DataTable)
		}
	case datatable.DateTypeKeywordTable:
		if v, ok := dh.keywordTableMap.Get(tag); ok {
			return v.(*datatable.DataTable)
		}
	case datatable.DateTypeDomainlistTable:
		if v, ok := dh.domainTableMap.Get(tag); ok {
			return v.(*datatable.DataTable)
		}
	}
	return nil
}

// 根据 tag 从 geoip.dat 加载 geoip 数据
func (dh *Datahub) reloadGeoipNetListByTag(tags []string, cache bool) error {
	if !cache {
		loader.RemoveCache(dh.geoipPath)
	}
	tagitems, err := loader.LoadGeoIPFromDATByTags(dh.geoipPath, tags)
	if err != nil {
		return err
	}
	dh.geonlmLock.Lock()
	defer dh.geonlmLock.Unlock()
	for _, dataitems := range tagitems {
		var nets []iplib.Net
		for _, data := range dataitems.GetCidr() {
			_net := iplib.NewNet(data.GetIp(), int(data.GetPrefix()))
			nets = append(nets, _net)
		}

		dh.geoipNetListMap[dataitems.GetCountryCode()] = netutils.NewNetList(nets)

	}
	return nil
}

// 根据 tag 从 geosite.dat 加载 geosite 数据
func (dh *Datahub) reloadGeositeDmoainListByTag(tags []string, cache bool) error {
	if !cache {
		loader.RemoveCache(dh.geositePath)
	}
	tagitems, err := loader.LoadGeoSiteFromDATByTags(dh.geositePath, tags)
	if err != nil {
		return err
	}
	dh.geodlmLock.Lock()
	defer dh.geodlmLock.Unlock()
	for _, dataitems := range tagitems {
		var sites []string
		var regexs []string
		for _, data := range dataitems.GetDomain() {
			switch data.Type {
			case v2data.Domain_Full, v2data.Domain_Domain:
				sites = append(sites, data.GetValue())
			case v2data.Domain_Regex:
				regexs = append(regexs, data.GetValue())
			}
		}
		dmlist := netutils.NewDomainList()
		dmlist.InitDomainData(netutils.MatchFullType, sites)
		dmlist.InitDomainData(netutils.MatchRegexType, regexs)
		dh.geositeDoaminListMap[dataitems.GetCountryCode()] = dmlist
	}
	return nil
}

func (dh *Datahub) parseDataTableByTag(datatype string, tags []string, from string) {
	for _, tag := range tags {
		tag = strings.ToUpper(tag)
		switch datatype {
		case datatable.DateTypeKeywordTable:
			table := datatable.NewFromArgs(datatable.DateTypeKeywordTable, tag, from)
			table.SetJwtSecret(dh.jwtSecret)
			table.SetBootstrap(dh.bootstrap)
			table.LoadAll()
			dh.keywordTableMap.Set(tag, table)
		case datatable.DateTypeDomainlistTable:
			table := datatable.NewFromArgs(datatable.DateTypeDomainlistTable, tag, from)
			table.SetJwtSecret(dh.jwtSecret)
			table.SetBootstrap(dh.bootstrap)
			table.LoadAll()
			dh.domainTableMap.Set(tag, table)
		case datatable.DateTypeNetlistTable:
			table := datatable.NewFromArgs(datatable.DateTypeNetlistTable, tag, from)
			table.SetJwtSecret(dh.jwtSecret)
			table.SetBootstrap(dh.bootstrap)
			table.LoadAll()
			dh.netlistTableMap.Set(tag, table)
		case datatable.DateTypeEcsTable:
			table := datatable.NewFromArgs(datatable.DateTypeEcsTable, tag, from)
			table.SetJwtSecret(dh.jwtSecret)
			table.SetBootstrap(dh.bootstrap)
			table.LoadAll()
			dh.ecsTableMap.Set(tag, table)
		}
	}

}

func (dh *Datahub) OnStartup() error {
	go func() {
		panic(dh.pubserver.start())
	}()
	log.Infof("pubserver is running %s", dh.pubserver.listenAddr)
	dh.startSched()
	log.Infof("sched is running")
	return nil
}

func (dh *Datahub) OnShutdown() error {
	err := dh.pubserver.stop()
	if err != nil {
		return err
	}
	dh.stopSched()
	return nil
}
