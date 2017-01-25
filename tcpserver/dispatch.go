package tcpserver

import (
	"fmt"
	"os"

	"github.com/zheng-ji/goSnowFlake"
)

//消息逻辑处理层，存储消息，分发消息，离线消息发送到push
type Dispatch struct {
	iw      *goSnowFlake.IdWorker
	sub     *Subscribe
	outChan chan *Packet
	quit    chan bool
}

//0 < workerId < 1024
func NewDispatch(workerId int64, nsqaddr string) *Dispatch {
	iw, err := goSnowFlake.NewIdWorker(workerId)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	d := &Dispatch{
		iw:      iw,
		outChan: make(chan *Packet, 1024),
		quit:    make(chan bool),
	}
	d.sub = NewSubscribe(&CustomProto{}, nsqaddr, MESSAGE_TOPIC_LOGIC, MESSAGE_CHANNEL_LOGIC_IM, d.outChan)

	return d
}

func (d *Dispatch) Run() {
	for {
		select {
		case p := <-d.outChan:
			d.handle(p)
		case <-d.quit:
			return
		}
	}
}

//处理消息分发
func (d *Dispatch) handle(p *Packet) {
	switch p.Mt {
	case MESSAGE_TYPE_P2P:
		//单聊
		d.handleP2p(p)
	case MESSAGE_TYPE_GROUP:
		//群消息
		d.handleGroup(p)
	case MESSAGE_TYPE_ROOM:
		//聊天室消息
		d.handleRoom(p)
	default:
		fmt.Printf("unknown message type: %d\n", p.Mt)
	}
}

func (d *Dispatch) handleP2p(p *Packet) {
	if id, err := d.iw.NextId(); err != nil {
		fmt.Println(err)
		return
	} else {
		p.Mid = id
	}

	d.sub.Publish(MESSAGE_TOPIC_DISPATCH, p)
}

func (d *Dispatch) handleGroup(p *Packet) {
	//获取群成员
	members := []int64{}
	for _, member := range members {
		if member == p.Sid {
			continue
		}

		packet := &Packet{
			Ver: p.Ver,
			Mt:  p.Mt,
			Mid: 0,
			Sid: p.Rid,
			Rid: member,
			Ext: p.Ext,
			Pl:  p.Pl,
		}
		if id, err := d.iw.NextId(); err != nil {
			fmt.Println(err)
			return
		} else {
			packet.Mid = id
		}

		d.sub.Publish(MESSAGE_TOPIC_DISPATCH, packet)
	}
}

func (d *Dispatch) handleRoom(p *Packet) {
	//获取聊天室成员
	members := []int64{}
	for _, member := range members {
		if member == p.Sid {
			continue
		}

		packet := &Packet{
			Ver: p.Ver,
			Mt:  p.Mt,
			Mid: 0,
			Sid: p.Rid,
			Rid: member,
			Ext: p.Ext,
			Pl:  p.Pl,
		}
		if id, err := d.iw.NextId(); err != nil {
			fmt.Println(err)
			return
		} else {
			packet.Mid = id
		}

		d.sub.Publish(MESSAGE_TOPIC_DISPATCH, packet)
	}
}

func (d *Dispatch) Close() {
	d.quit <- true
}
