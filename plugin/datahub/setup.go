package datahub

import (
	"strings"

	"github.com/ca17/dnssrc/plugin/pkg/validutil"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
)

var log = clog.NewWithPlugin("datahub")

func init() { plugin.Register("datahub", setup) }

func setup(c *caddy.Controller) error {
	datahub, err := parseConfig(c)
	if err != nil {
		return plugin.Error("datahub", err)
	}

	if datahub.debug {
		datahub.debugPrint()
	}

	c.OnStartup(func() error {
		return datahub.OnStartup()
	})

	c.OnFinalShutdown(func() error {
		return datahub.OnShutdown()
	})

	dnsserver.GetConfig(c).AddPlugin(
		func(next plugin.Handler) plugin.Handler {
			datahub.Next = next
			return datahub
		})

	return nil
}

func parseConfig(c *caddy.Controller) (*Datahub, error) {
	d := NewDatahub()
	i := 0
	for c.Next() {
		if i > 0 {
			return nil, plugin.ErrOnce
		}
		i++
		for c.NextBlock() {
			switch c.Val() {
			case "geoip_path":
				remaining := c.RemainingArgs()
				plen := len(remaining)
				if plen < 1 {
					return nil, c.Errf("geoip_path format is `geoip_path  filepath` ")
				}
				d.geoipPath = remaining[0]
			case "geosite_path":
				remaining := c.RemainingArgs()
				plen := len(remaining)
				if plen < 1 {
					return nil, c.Errf("geosite_path format is `geosite_path filepath` ")
				}
				d.geositePath = remaining[0]
			case "geoip_cache":
				d.geoipCacheTags = c.RemainingArgs()
				plen := len(d.geoipCacheTags)
				if plen < 1 {
					return nil, c.Errf("geoip_cache format is `geoip_cache tag...` ")
				}
				err := d.reloadGeoipNetListByTag(d.geoipCacheTags, true)
				if err != nil {
					return nil, c.Errf("load geoip.dat error")
				}
			case "geosite_cache":
				d.geositeCacheTags = c.RemainingArgs()
				plen := len(d.geositeCacheTags)
				if plen < 1 {
					return nil, c.Errf("geosite_cache format is `geosite_cache tag...` ")
				}
				err := d.reloadGeositeDmoainListByTag(d.geositeCacheTags, true)
				if err != nil {
					return nil, c.Errf("load geosite.dat error")
				}
			case "geodat_upgrade_url":
				remaining := c.RemainingArgs()
				plen := len(remaining)
				if plen < 1 {
					return nil, c.Errf("geodat_upgrade_url no value ")
				}
				if !validutil.IsURL(remaining[0]) {
					return nil, c.Errf("geodat_upgrade_url format must url ")
				}
				d.geodatUpgradeUrl = remaining[0]
			case "geodat_upgrade_cron":
				cronSpec := strings.Join(c.RemainingArgs(), " ")
				_, err := cronParser.Parse(cronSpec)
				if err != nil {
					return nil, c.Errf("geodat_upgrade_cron format must cron string ")
				}
				d.geodatUpgradeCron = cronSpec
			case "keyword_table":
				remaining := c.RemainingArgs()
				plen := len(remaining)
				if plen != 2 {
					return nil, c.Errf("keyword_table args num is 2 ")
				}
				err := d.parseKeywordTableByTag(remaining[0], remaining[1])
				if err != nil {
					return nil, c.Errf("parseKeywordTableByTag  err %s", err.Error())
				}
			case "reload":
				reloadCron := strings.Join(c.RemainingArgs(), " ")
				_, err := cronParser.Parse(reloadCron)
				if err != nil {
					return nil, c.Errf("reload_cron format must cron string ")
				}
				d.reloadCron = reloadCron
			case "debug":
				d.debug = true
			default:

			}
		}
	}
	return d, nil
}
