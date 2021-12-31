package datahub

import (
	"github.com/ca17/datahub/plugin/pkg/datatable"
)

func (dh *Datahub) startSched() {

	_, _ = dh.sched.AddFunc(dh.geodatUpgradeCron, func() {
		_ = dh.reloadGeoipNetListByTag(dh.geoipCacheTags, false)
		_ = dh.reloadGeositeDmoainListByTag(dh.geoipCacheTags, false)
	})

	_, _ = dh.sched.AddFunc(dh.reloadCron, func() {
		dh.cronUpdateKeywordTableMap()
	})

	_, _ = dh.sched.AddFunc("@every 60s", func() {
		dh.dayaDomainChartStat.Update(dh.domainMatchStat)
		dh.dayNetworkChartStat.Update(dh.networkMatchStat)
	})

	dh.sched.Start()
}

func (dh *Datahub) stopSched() {
	dh.sched.Stop()
}

func (dh *Datahub) cronUpdateKeywordTableMap() {
	for _, _item := range dh.keywordTableMap.Items() {
		item := _item.(*datatable.DataTable)
		item.LoadFromFile()
		if dh.jwtSecret != "" {
			item.LoadFromUrl()
		} else {
			item.LoadFromUrl()
		}
	}
}
