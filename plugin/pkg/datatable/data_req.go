package datatable

// DataReq 重载数据body参数
type DataReq struct {
	Domain  bool `json:"domain"`
	Network bool `json:"network"`
	Ecs     bool `json:"ecs"`
	Keyword bool `json:"keyword"`
}
