package datahub

import (
	"bufio"
	"github.com/ca17/datahub/plugin/pkg/common"
	"github.com/ca17/datahub/plugin/pkg/httpc"
	"github.com/ca17/dnssrc/plugin/pkg/validutil"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/dnstap"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"net"
	"os"
	"strings"
)

var log = clog.NewWithPlugin("datahub")

func init() { plugin.Register("datahub", setup) }

var tapPlugin *dnstap.Dnstap

func setup(c *caddy.Controller) error {
	datahub, err := parseConfig(c)
	if err != nil {
		return plugin.Error("datahub", err)
	}

	c.OnStartup(func() error {
		if dtap := dnsserver.GetConfig(c).Handler("dnstap"); dtap != nil {
			if hp, ok := dtap.(*dnstap.Dnstap); ok {
				tapPlugin = hp
			}
		}
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
			switch dir := c.Val(); dir {
			case "bootstrap":
				remaining := c.RemainingArgs()
				plen := len(remaining)
				if plen == 0 {
					return nil, c.ArgErr()
				}
				d.bootstrap = remaining
				log.Infof("%s: %v", dir, d.bootstrap)
			case "geoip_path":
				remaining := c.RemainingArgs()
				plen := len(remaining)
				if plen < 1 {
					return nil, c.Errf("geoip_path format is `geoip_path  filepath` ")
				}
				d.geoipPath = remaining[0]
				log.Info("geoip_path ", d.geoipPath)
			case "geosite_path":
				remaining := c.RemainingArgs()
				plen := len(remaining)
				if plen < 1 {
					return nil, c.Errf("geosite_path format is `geosite_path filepath` ")
				}
				d.geositePath = remaining[0]
				log.Info("geosite_path ", d.geositePath)
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
				log.Info("geoip_cache ", d.geoipCacheTags)
				for k, v := range d.geoipNetListMap {
					log.Infof("geoip_cache %s total %d", k, v.Len())
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
				log.Info("geosite_cache ", d.geositeCacheTags)
				for k, v := range d.geositeDoaminListMap {
					log.Infof("geosite_cache %s full_domain:%d regex_domain:%d", k, v.FullLen(), v.RegexLen())
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
				log.Info("geodat_upgrade_url ", d.geodatUpgradeUrl)
			case "geodat_upgrade_cron":
				cronSpec := strings.Join(c.RemainingArgs(), " ")
				_, err := cronParser.Parse(cronSpec)
				if err != nil {
					return nil, c.Errf("geodat_upgrade_cron format must cron string ")
				}
				d.geodatUpgradeCron = cronSpec
				log.Info("geodat_upgrade_cron ", d.geodatUpgradeCron)
			case "datatables": // parse datatables
				remaining := c.RemainingArgs()
				plen := len(remaining)
				if plen != 1 {
					return nil, c.ArgErr()
				}
				err := parseDataTables(d, remaining[0])
				if err != nil {
					return nil, err
				}
				// TODO setup阶段xxx_table存在则更新table
			//case "keyword_table":
			//	remaining := c.RemainingArgs()
			//	plen := len(remaining)
			//	if plen != 2 {
			//		return nil, c.ArgErr()
			//	}
			//	d.parseDataTableByTag(datatable.DateTypeKeywordTable, strings.Split(remaining[0], ","), remaining[1])
			//	d.keywordTableMap.IterCb(func(k string, v interface{}) {
			//		log.Infof("keyword_table %s total %d", k, v.(*datatable.DataTable).Len())
			//	})
			//case "domain_table":
			//	remaining := c.RemainingArgs()
			//	plen := len(remaining)
			//	if plen != 2 {
			//		return nil, c.Errf("domain_table args num is 2 ")
			//	}
			//	d.parseDataTableByTag(datatable.DateTypeDomainlistTable, strings.Split(remaining[0], ","), remaining[1])
			//	d.domainTableMap.IterCb(func(k string, v interface{}) {
			//		log.Infof("domain_table %s total %d", k, v.(*datatable.DataTable).Len())
			//	})
			//case "netlist_table":
			//	remaining := c.RemainingArgs()
			//	plen := len(remaining)
			//	if plen != 2 {
			//		return nil, c.ArgErr()
			//	}
			//	d.parseDataTableByTag(datatable.DateTypeNetlistTable, strings.Split(remaining[0], ","), remaining[1])
			//	d.netlistTableMap.IterCb(func(k string, v interface{}) {
			//		log.Infof("netlist_table %s total %d", k, v.(*datatable.DataTable).Len())
			//	})
			//case "ecs_table":
			//	remaining := c.RemainingArgs()
			//	plen := len(remaining)
			//	if plen != 2 {
			//		return nil, c.ArgErr()
			//	}
			//	d.parseDataTableByTag(datatable.DateTypeEcsTable, strings.Split(remaining[0], ","), remaining[1])
			//	d.ecsTableMap.IterCb(func(k string, v interface{}) {
			//		log.Infof("ecs_table %s total %d", k, v.(*datatable.DataTable).Len())
			//	})
			case "datapub_listen":
				remaining := c.RemainingArgs()
				plen := len(remaining)
				if plen < 1 {
					return nil, c.ArgErr()
				}
				_, _, err := net.SplitHostPort(remaining[0])
				if err != nil {
					return nil, c.SyntaxErr(err.Error())
				}
				d.pubserver.listenAddr = remaining[0]
				if plen == 3 {
					d.pubserver.certfile = remaining[1]
					d.pubserver.keyfile = remaining[2]
				}
			case "notify_server":
				remaining := c.RemainingArgs()
				plen := len(remaining)
				if plen < 1 {
					return nil, c.ArgErr()
				}
				if d.notifyServer.addServer(remaining...) {
					log.Infof("add notify server %v", remaining)
				}
			case "jwt_secret":
				remaining := c.RemainingArgs()
				plen := len(remaining)
				if plen < 1 {
					return nil, c.ArgErr()
				}
				d.jwtSecret = remaining[0]
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
				log.Errorf("Unsupported directives %s", dir)
			}
		}
	}
	return d, nil
}

// parseDataTables 解析 xx_tables数据
func parseDataTables(d *Datahub, path string) error {
	// 解析datatables
	var lines []string
	switch {
	case validutil.IsURL(path):
		res, err := httpc.Get(path, nil, d.bootstrap, 30)
		if err != nil {
			log.Errorf("load datatables from url[%s] error %s", path, err.Error())
			return err
		}
		lines = strings.Split(string(res), "\n")
	case common.FileExists(path):
		file, err := os.Open(path)
		if err != nil {
			log.Errorf("load datatables error %s", err.Error())
			return err
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			lines = append(lines, line)
		}
	default:
		log.Infof("load datatables, path[%s] cloud not match type[url, file]", path)
		return nil
	}

	for _, line := range lines {
		attrs := strings.Fields(line)
		if len(attrs) != 3 {
			log.Errorf("format error: %s", line)
			continue
		}
		datatype, tags, subfrom := attrs[0], attrs[1], attrs[2]
		args := strings.Split(tags, ",")
		d.parseDataTableByTag(datatype, args, subfrom)
	}
	return nil
}
