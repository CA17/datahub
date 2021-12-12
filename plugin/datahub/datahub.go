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
	"github.com/ca17/datahub/plugin/pkg/v2data"
	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
	"github.com/robfig/cron/v3"
)

var cronParser = cron.NewParser(
	cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

type Datahub struct {
	nlmLock              sync.RWMutex
	dlmLock              sync.RWMutex
	ktLock               sync.RWMutex
	Next                 plugin.Handler
	geoipCacheTags       []string
	geositeCacheTags     []string
	geoipNetListMap      map[string]*netutils.NetList
	geositeDoaminListMap map[string]*netutils.DomainList
	keywordTableMap      map[string]*datatable.DataTable
	geoipPath            string
	geositePath          string
	geodatUpgradeUrl     string
	geodatUpgradeCron    string
	sched                *cron.Cron
	matchCache           *bigcache.BigCache
	reloadCron           string
	debug                bool
}

func (dh *Datahub) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	return plugin.NextOrFailure(dh.Name(), dh.Next, ctx, w, r)
}

func (dh *Datahub) Name() string { return "datahub" }

func NewDatahub() *Datahub {
	mc, _ := bigcache.NewBigCache(bigcache.DefaultConfig(time.Second * 3600))
	return &Datahub{
		nlmLock:              sync.RWMutex{},
		dlmLock:              sync.RWMutex{},
		geoipNetListMap:      make(map[string]*netutils.NetList),
		geositeDoaminListMap: make(map[string]*netutils.DomainList),
		keywordTableMap:      make(map[string]*datatable.DataTable),
		matchCache:           mc,
		sched:                cron.New(cron.WithParser(cronParser)),
	}
}

func (dh *Datahub) getDomainListByTag(tag string) *netutils.DomainList{
	dh.dlmLock.RLock()
	defer dh.dlmLock.RUnlock()
	return dh.geositeDoaminListMap[tag]
}

func (dh *Datahub) getNetListByTag(tag string) *netutils.NetList{
	dh.nlmLock.RLock()
	defer dh.nlmLock.RUnlock()
	return dh.geoipNetListMap[tag]
}

func (dh *Datahub) getKeywordTableByTag(tag string) *datatable.DataTable{
	dh.ktLock.RLock()
	defer dh.ktLock.RUnlock()
	return dh.keywordTableMap[tag]
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
	dh.nlmLock.Lock()
	defer dh.nlmLock.Unlock()
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
	dh.dlmLock.Lock()
	defer dh.dlmLock.Unlock()
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

func (dh *Datahub) parseKeywordTableByTag(tag string, from string) error {
	tag = strings.ToUpper(tag)
	table, err := datatable.NewFromArgs(datatable.DateTypeKeywordTable, tag, from )
	if err != nil {
		return err
	}
	table.LoadAll()
	dh.ktLock.Lock()
	defer dh.ktLock.Unlock()
	dh.keywordTableMap[tag] = table
	return nil
}

func (dh *Datahub) OnStartup() error {
	dh.startSched()
	return nil
}

func (dh *Datahub) OnShutdown() error {
	dh.stopSched()
	return nil
}

func (dh *Datahub) debugPrint() {
	log.Info("geoip_path ", dh.geositePath)
	log.Info("geosite_path ", dh.geositePath)
	log.Info("geoip_cache ", dh.geoipCacheTags)
	log.Info("geosite_cache ", dh.geositeCacheTags)
	log.Info("geodat_upgrade_url ", dh.geodatUpgradeUrl)
	log.Info("geodat_upgrade_cron ", dh.geodatUpgradeCron)
	for k, v := range dh.geoipNetListMap {
		log.Infof("geoip_cache %s total %d", k, v.Len())
	}
	for k, v := range dh.geositeDoaminListMap {
		log.Infof("geosite_cache %s full_domain:%d regex_domain:%d", k, v.FullLen(), v.RegexLen())
	}
	for k, v := range dh.keywordTableMap {
		log.Infof("keyword_table %s total %d",k, v.Len())
	}
}
