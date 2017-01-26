package tcpserver

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

//推送消息服务
type PushSrv struct {
	pool    *redis.Pool
	sub     *Subscribe
	outChan chan *Packet
	quit    chan bool
}

func NewPushSrv(nsqaddr string, host string, pwd string, db int) *PushSrv {
	ps := &PushSrv{
		outChan: make(chan *Packet, 1024),
		quit:    make(chan bool),
		pool: &redis.Pool{
			MaxIdle:     5,
			IdleTimeout: 300 * time.Second,
			// Other pool configuration not shown in this example.
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", host)
				if err != nil {
					return nil, err
				}
				if _, err := c.Do("AUTH", pwd); err != nil {
					c.Close()
					return nil, err
				}
				if _, err := c.Do("SELECT", db); err != nil {
					c.Close()
					return nil, err
				}
				return c, nil
			},
		},
	}

	ps.sub = NewSubscribe(&CustomProto{}, nsqaddr, MESSAGE_TOPIC_DISPATCH, MESSAGE_CHANNEL_DISPATCH_PUSH, ps.outChan)

	return ps
}

func (ps *PushSrv) Run() {
	for {
		select {
		case p := <-ps.outChan:
			ps.handle(p)
		case <-ps.quit:
			return
		}
	}
}

func (ps *PushSrv) handle(p *Packet) {
	switch p.Mt {
	case MESSAGE_TYPE_P2P:
		//单聊
		ps.handleP2p(p)
	case MESSAGE_TYPE_GROUP:
		//群消息
		ps.handleGroup(p)
	case MESSAGE_TYPE_ROOM:
		//聊天室消息
		ps.handleRoom(p)
	default:
		fmt.Printf("unknown message type: %d\n", p.Mt)
	}
}

func (ps *PushSrv) handleP2p(p *Packet) {

}

func (ps *PushSrv) handleGroup(p *Packet) {

}

func (ps *PushSrv) handleRoom(p *Packet) {

}

func (ps *PushSrv) Close() {
	ps.quit <- true
}
