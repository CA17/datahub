package datahub

import (
	"github.com/coredns/coredns/request"
)

type DnsNotify struct {
	Topic  string `json:"topic"`
	Client string `json:"client"`
	QName  string `json:"qname"`
	Class  string `json:"class"`
	QType  string `json:"qtype"`
}

func (dh *Datahub) NotifyMessage(topic string, state *request.Request) {
	nmsg := new(DnsNotify)
	nmsg.Topic = topic
	nmsg.Client = state.IP()
	nmsg.QName = state.QName()
	nmsg.QType = state.Type()
	nmsg.Class = state.Class()
	go dh.notifyServer.sendNotify(topic, dh.bootstrap, nmsg)
}
