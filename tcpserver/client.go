package tcpserver

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

var (
	MESSAGE_TYPE_HEARTBEAT       = int32(1)  //心跳消息
	MESSAGE_TYPE_PING            = int32(2)  //ping
	MESSAGE_TYPE_PONG            = int32(3)  //pong
	MESSAGE_TYPE_REGISTER        = int32(4)  //注册设备token
	MESSAGE_TYPE_REGISTER_STATUS = int32(5)  //注册设备token回执
	MESSAGE_TYPE_AUTH            = int32(6)  //鉴权消息
	MESSAGE_TYPE_AUTH_STATUS     = int32(7)  //鉴权回执
	MESSAGE_TYPE_P2P             = int32(8)  //单聊消息
	MESSAGE_TYPE_ACK             = int32(9)  //消息ack回执
	MESSAGE_TYPE_GROUP           = int32(10) //群聊消息
	MESSAGE_TYPE_ROOM            = int32(11) //聊天室消息
)

type DeviceInfo struct {
	token string
}

type AuthInfo struct {
	uid   int64
	token string
}

type ClientCallback interface {
	OnConnect() bool
	OnAuth() bool
	OnMessage(*Packet) bool
	OnClose() bool
}

type Client struct {
	uid         int64  //用户id作为客户端唯一id
	deviceToken string //用户设备token作为客户端唯一id
	server      *TCPServer
	conn        net.Conn
	sendChan    chan *Packet //发送数据到客户端
	receiveChan chan *Packet //从客户端接收数据
	quit        chan struct{}
	authFlag    int32
	closeFlag   int32
	closeOnce   sync.Once //保证调用一次close
}

func NewClient(s *TCPServer, c net.Conn) *Client {
	return &Client{
		server:      s,
		conn:        c,
		sendChan:    make(chan *Packet, 1024),
		receiveChan: make(chan *Packet, 1024),
		quit:        make(chan struct{}),
	}
}

func (client *Client) OnConnect() bool {
	fmt.Println("connect success")
	return true
}

func (client *Client) OnAuth() bool {
	return true
}

//根据收到的消息类型，不同的处理逻辑
func (client *Client) OnMessage(p *Packet) bool {
	switch p.mt {
	case MESSAGE_TYPE_HEARTBEAT:
		//心跳
		client.handleHeartbeat(p)
	case MESSAGE_TYPE_PING:
		//ping
		client.handlePing(p)
	case MESSAGE_TYPE_REGISTER:
		//注册设备信息
		client.handleRegister(p)
	case MESSAGE_TYPE_AUTH:
		//鉴权
		client.handleAuth(p)
	case MESSAGE_TYPE_P2P:
		//单聊
		client.handleP2p(p)
	case MESSAGE_TYPE_GROUP:
		//群消息
		client.handleGroup(p)
	case MESSAGE_TYPE_ROOM:
		//聊天室消息
		client.handleRoom(p)
	default:
		fmt.Printf("unknown message type: %d\n", p.mt)
	}

	return true
}

func (client *Client) handleHeartbeat(p *Packet) {

}

func (client *Client) handlePing(p *Packet) {

}

func (client *Client) handleRegister(p *Packet) {
	//从payload获取注册设备信息
	deviceInfo := DeviceInfo{}
	err := json.Unmarshal(p.pl, &deviceInfo)
	if err != nil || deviceInfo.token == "" {
		//输出注册失败信息

		return
	}

	c := client.server.GetClientByDt(deviceInfo.token)
	if c == nil {
		//注册新的客户端
		client.server.RegisterClientByDt(client, deviceInfo.token)
	} else {
		//有老的客户端
		if client == c {
			//同一个客户端 do nothing
		} else {
			//不是同一个客户端，注销之前的客户端
			client.server.UnRegisterClient(c.uid, deviceInfo.token)
			//注册新的客户端
			client.server.RegisterClientByDt(client, deviceInfo.token)
		}
	}

	if client.IsAuth() {
		client.server.RegisterClientByUid(client, client.uid)
	}

	//返回成功回执
}

func (client *Client) handleAuth(p *Packet) {
	//获取鉴权信息
	authInfo := AuthInfo{}
	err := json.Unmarshal(p.pl, &authInfo)
	if err != nil || authInfo.uid == 0 || authInfo.token == "" {
		//输出鉴权失败信息

		return
	}

	//获取鉴权信息
	//判断鉴权通过

	//通过
	c := client.server.GetClientByUid(authInfo.uid)
	if c == nil {
		//注册新的客户端
		client.server.RegisterClientByUid(client, authInfo.uid)
	} else {
		//有老的客户端
		if client == c {
			//同一个客户端 do nothing
		} else {
			//不是同一个客户端，注销之前的客户端
			client.server.UnRegisterClient(authInfo.uid, c.deviceToken)
			//注册新的客户端
			client.server.RegisterClientByDt(client, client.deviceToken)
			client.server.RegisterClientByUid(client, authInfo.uid)
		}
	}

	atomic.StoreInt32(&client.authFlag, 1)

	//成功回执
}

func (client *Client) handleP2p(p *Packet) {
	if !client.IsAuth() {
		//没有通过鉴权
		return
	}

	client.server.inChan <- p
}

func (client *Client) handleGroup(p *Packet) {
	if !client.IsAuth() {
		//没有通过鉴权
		return
	}

	client.server.inChan <- p
}

func (client *Client) handleRoom(p *Packet) {
	if !client.IsAuth() {
		//没有通过鉴权
		return
	}

	client.server.inChan <- p
}

func (client *Client) OnClose() bool {
	fmt.Println("connect close success")
	return true
}

func (client *Client) Close() {
	client.closeOnce.Do(func() {
		client.server.UnRegisterClient(client.uid, client.deviceToken)
		atomic.StoreInt32(&client.closeFlag, 1) //标记关闭
		atomic.StoreInt32(&client.authFlag, 0)  //标记关闭
		close(client.quit)
		close(client.sendChan)
		close(client.receiveChan)
		client.conn.Close()
		client.OnClose()
	})
}

func (client *Client) IsClose() bool {
	return atomic.LoadInt32(&client.closeFlag) == 1
}

func (client *Client) IsAuth() bool {
	return atomic.LoadInt32(&client.authFlag) == 1
}

func (client *Client) Do() {
	if !client.OnConnect() {
		return
	}

	go client.readLoop()
	go client.writeLoop()
	go client.handleLoop()
}

//读取来自客户端数据，按照protocol协议解析packet
func (client *Client) readLoop() {
	defer func() {
		recover()
		client.Close()
	}()

	for {
		select {
		case <-client.server.quit:
			return
		case <-client.quit:
			return
		default:
		}

		//读取数据
		p, err := client.server.protocol.ReadPacket(client.conn)
		if err != nil {
			fmt.Printf("read packet error: %s\n", err.Error())
			return
		}

		client.receiveChan <- p
	}
}

//写入数据到客户端
func (client *Client) writeLoop() {
	defer func() {
		recover()
		client.Close()
	}()

	for {
		select {
		case <-client.server.quit:
			return
		case <-client.quit:
			return
		case p := <-client.sendChan:
			if client.IsClose() {
				return
			}

			err := client.server.protocol.WritePacket(client.conn, p)
			if err != nil {
				return
			}
		}
	}
}

func (client *Client) handleLoop() {
	defer func() {
		recover()
		client.Close()
	}()

	for {
		select {
		case <-client.server.quit:
			return
		case <-client.quit:
			return
		case p := <-client.receiveChan:
			if client.IsClose() {
				return
			}

			if !client.OnMessage(p) {
				return
			}
		}
	}
}
