package tcpserver

import "fmt"

//存储消息服务
type StoreSrv struct {
	sub     *Subscribe
	outChan chan *Packet
	quit    chan bool
}

func NewStoreSrv(nsqaddr string) *StoreSrv {
	ps := &StoreSrv{
		outChan: make(chan *Packet, 1024),
		quit:    make(chan bool),
	}

	ps.sub = NewSubscribe(&CustomProto{}, nsqaddr, MESSAGE_TOPIC_DISPATCH, MESSAGE_CHANNEL_DISPATCH_STORE, ps.outChan)

	return ps
}

func (ss *StoreSrv) Run() {
	for {
		select {
		case p := <-ss.outChan:
			ss.handle(p)
		case <-ss.quit:
			return
		}
	}
}

func (ss *StoreSrv) handle(p *Packet) {
	switch p.Mt {
	case MESSAGE_TYPE_P2P:
		//单聊
		ss.handleP2p(p)
	case MESSAGE_TYPE_GROUP:
		//群消息
		ss.handleGroup(p)
	case MESSAGE_TYPE_ROOM:
		//聊天室消息
		ss.handleRoom(p)
	default:
		fmt.Printf("unknown message type: %d\n", p.Mt)
	}
}

func (ss *StoreSrv) handleP2p(p *Packet) {

}

func (ss *StoreSrv) handleGroup(p *Packet) {

}

func (ss *StoreSrv) handleRoom(p *Packet) {

}

func (ss *StoreSrv) Close() {
	ss.quit <- true
}
