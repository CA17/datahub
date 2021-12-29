package datahub

import (
	"encoding/json"
	"net/http"

	"github.com/c-robinson/iplib"
	"github.com/ca17/datahub/plugin/pkg/common"
	"github.com/ca17/datahub/plugin/pkg/datatable"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

const (
	charsetUTF8                          = "charset=UTF-8"
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + charsetUTF8
	MIMETextXML                          = "text/xml"
	MIMETextXMLCharsetUTF8               = MIMETextXML + "; " + charsetUTF8
	MIMEApplicationForm                  = "application/x-www-form-urlencoded"
	MIMEApplicationProtobuf              = "application/protobuf"
	MIMEApplicationMsgpack               = "application/msgpack"
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm                    = "multipart/form-data"
	MIMEOctetStream                      = "application/octet-stream"
)

type dataServer struct {
	listenAddr string
	certfile   string
	keyfile    string
	server     *fasthttp.Server
	router     *routing.Router
	hub        *Datahub
}

func newPubServer(hub *Datahub) *dataServer {
	return &dataServer{router: routing.New(), hub: hub}
}

func (s *dataServer) fetchDataTable(c *routing.Context, datatype string, tag string, limit int) {
	if list := s.hub.getDataTableByTag(datatype, tag); list != nil {
		list.GetData().ForEach(func(item interface{}) error {
			switch item.(type) {
			case iplib.Net:
				i, _ := c.WriteString(item.(iplib.Net).String())
				if i > 0 {
					_, _ = c.WriteString("\n")
				}
			case string:
				i, _ := c.WriteString(item.(string))
				if i > 0 {
					_, _ = c.WriteString("\n")
				}
			}
			return nil
		}, limit)
	}
}

func (s *dataServer) listNetBytag(c *routing.Context) error {
	tag := c.Param("tag")
	if tag == "" {
		c.Error("tag is empty", http.StatusBadRequest)
	}
	limit, err := c.QueryArgs().GetUint("limit")
	if err != nil {
		limit = 1000
	}
	s.fetchDataTable(c, datatable.DateTypeNetlistTable, tag, limit)
	return nil
}

func (s *dataServer) listDomainBytag(c *routing.Context) error {
	tag := c.Param("tag")
	if tag == "" {
		c.Error("tag is empty", http.StatusBadRequest)
	}
	limit, err := c.QueryArgs().GetUint("limit")
	if err != nil {
		limit = 1000
	}
	s.fetchDataTable(c, datatable.DateTypeDomainlistTable, tag, limit)
	return nil
}

func (s *dataServer) listKeywordsBytag(c *routing.Context) error {
	tag := c.Param("tag")
	if tag == "" {
		c.Error("tag is empty", http.StatusBadRequest)
	}
	limit, err := c.QueryArgs().GetUint("limit")
	if err != nil {
		limit = 1000
	}
	s.fetchDataTable(c, datatable.DateTypeKeywordTable, tag, limit)
	return nil
}

// reloadData 重新加载网络地址，域名， 关键词， ecs等信息
func (s *dataServer) reloadData(c *routing.Context) error {
	body := c.Request.Body()
	var dataReq datatable.DataReq
	err := json.Unmarshal(body, &dataReq)
	if err != nil {
		// 设置状态为服务器错误编码
		c.SetStatusCode(500)
		_, err = c.WriteString(err.Error())
		return err
	}
	// 重载数据domain
	if dataReq.Domain {
		go func() {
			domainMap := s.hub.domainTableMap
			if len(domainMap) > 0 {
				domainMap.IterCb(func(k string, v interface{}) {
					v.(*datatable.DataTable).LoadAll()
				})
			}
		}()
	}

	// 重载数据network
	if dataReq.Network {
		go func() {
			networkMap := s.hub.netlistTableMap
			if len(networkMap) > 0 {
				networkMap.IterCb(func(k string, v interface{}) {
					v.(*datatable.DataTable).LoadAll()
				})
			}
		}()
	}

	// 重载数据ecs
	if dataReq.Ecs {
		go func() {
			ecsMap := s.hub.ecsTableMap
			if len(ecsMap) > 0 {
				ecsMap.IterCb(func(k string, v interface{}) {
					v.(*datatable.DataTable).LoadAll()
				})
			}
		}()
	}

	// 重载数据keyword
	if dataReq.Keyword {
		go func() {
			keywordMap := s.hub.keywordTableMap
			if len(keywordMap) > 0 {
				keywordMap.IterCb(func(k string, v interface{}) {
					v.(*datatable.DataTable).LoadAll()
				})
			}
		}()
	}
	_, err = c.WriteString("ok")
	return err
}

func (s *dataServer) start() error {
	if s.router == nil {
		s.router = routing.New()
	}
	s.router.Get("/net/list/<tag>", s.listNetBytag)
	s.router.Get("/domain/list/<tag>", s.listDomainBytag)
	s.router.Get("/keyword/list/<tag>", s.listKeywordsBytag)
	s.router.Post("/reload/data", s.reloadData)
	s.server = &fasthttp.Server{
		Handler: s.router.HandleRequest,
	}
	if s.certfile != "" && common.FileExists(s.certfile) &&
		s.keyfile != "" && common.FileExists(s.keyfile) {
		return s.server.ListenAndServeTLS(s.listenAddr, s.certfile, s.keyfile)
	}
	return s.server.ListenAndServe(s.listenAddr)
}

func (s *dataServer) stop() error {
	return s.server.Shutdown()
}
