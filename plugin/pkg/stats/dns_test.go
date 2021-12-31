package stats

import (
	"testing"
	"time"

	"github.com/metaslink/metasdns/plugin/pkg/common"
)

func TestNewDayDnsStat(t *testing.T) {
	ds := NewDayDnsStat(time.Second * 15)
	c := NewCounterStat()
	start := time.Now()
	for {
		if time.Now().Sub(start).Seconds() > 12 {
			break
		}
		c.Incr("cn", 1)
		time.Sleep(time.Second * 3)
		c.Incr("cn", 1)
		c.Incr("ads", 1)
		t.Log("...")
		ds.Update(c)
	}
	ds.Rolling()
	t.Log(common.ToJson(ds.LineChartData("最近24小时统计").ChartData()))
}
