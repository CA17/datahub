package stats

import (
	"github.com/ahmetb/go-linq"
)

type LineChartData struct {
	Title     string             `json:"title"`
	YaxisName string             `json:"yaxis_name"`
	Times     []string           `json:"times"`
	Datas     map[string][]int64 `json:"datas"`
}

func NewLineChartData(title string, yaxisName string) *LineChartData {
	return &LineChartData{Title: title, YaxisName: yaxisName, Times: make([]string, 0), Datas: make(map[string][]int64, 0)}
}

func (l *LineChartData) ChartData() *LineChartData {
	linq.From(l.Datas).ForEachT(func(kv linq.KeyValue) {
		var last int64 = 0
		vals := make([]int64, 0)
		linq.From(kv.Value).ForEachT(func(v int64) {
			var val int64 = 0
			if last == 0 {
				val = 0
				last = v
			} else {
				val = v - last
				if val < 0 {
					val = 0
				}
				last = v
			}
			vals = append(vals, val)
		})
		l.Datas[kv.Key.(string)] = vals
	})
	return l
}
