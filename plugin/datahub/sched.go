package datahub

func (dh *Datahub) startSched() {

	_, _ = dh.sched.AddFunc(dh.geodatUpgradeCron, func() {
		_ = dh.reloadGeoipNetListByTag(dh.geoipCacheTags, false)
		_ = dh.reloadGeositeDmoainListByTag(dh.geoipCacheTags, false)
	})

	_, _ = dh.sched.AddFunc(dh.reloadCron, func() {
		dh.ktLock.Lock()
		defer dh.ktLock.Unlock()
		for _, item := range dh.keywordTableMap {
			item.LoadFromFile()
			item.LoadFromUrl()
		}
	})

	dh.sched.Start()
}

func (dh *Datahub) stopSched() {
	dh.sched.Stop()
}
