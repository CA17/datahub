package datahub

import (
	"github.com/ca17/datahub/plugin/pkg/httpc"
	"github.com/ca17/datahub/plugin/pkg/stringset"
	"github.com/ca17/datahub/plugin/pkg/validutil"
)

type notifyServer struct {
	servers *stringset.StringSet
}

func newNotifyServer() *notifyServer {
	return &notifyServer{servers: stringset.New()}
}

func (s *notifyServer) addServer(server ...string) bool {
	if !validutil.IsURL(server) {
		return false
	}
	s.servers.Add(server...)
	return true
}

func (s *notifyServer) sendNotify(topic string, nmsg *DnsNotify) {
	s.servers.ForEach(func(url string) {
		_, err := httpc.PostJson(url, nmsg)
		if err != nil {
			log.Errorf("send notify error %s", err.Error())
		}
	})
}
