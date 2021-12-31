# datahub

*datahub* - a CoreDNS Data Manage Plugin

##  相关项目

[]

## 配置案例

```
.:53 {
    datahub {
        bootstrap 114.114.114.114:53 # 引导 DNS
        jwt_secret 9b6de5cc-vcty-4bf1-zpms-0f568ac9da37
        geoip_path conf/geoip.dat
        geosite_path conf/geosite.dat
        geoip_cache cn hk jp google apple
        geosite_cache cn hk jp private apple
        geodat_upgrade_url http://teamsacs.mydomain.cn/geodat
        geodat_upgrade_cron 0 30 0 * * *
        # datatables conf/datatables.txt
        keyword_table  cn,google  conf/keywords.txt
        domain_table  cn,aliyun,ads  conf/domains.txt
        netlist_table  cn,aliyun,local,office  conf/networks.txt
        ecs_table  global  conf/ecs_table.txt
        datapub_listen :9800
        notify_server  https://teamsacs.appsway.cn
        reload @every 3s
    }
}

```