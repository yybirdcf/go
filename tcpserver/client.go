package tcpserver

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

type ResponseInfo struct {
	Status int64
	Msg    string
}

type DeviceInfo struct {
	Token string
}

type AuthInfo struct {
	Uid   int64
	Token string
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

func buildResponseInfo(status int64, msg string) []byte {
	resp := ResponseInfo{
		Status: status,
		Msg:    msg,
	}

	data, _ := json.Marshal(resp)
	return data
}

//根据收到的消息类型，不同的处理逻辑
func (client *Client) OnMessage(p *Packet) bool {
	switch p.Mt {
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
		fmt.Printf("unknown message type: %d\n", p.Mt)
	}

	return true
}

func (client *Client) handleHeartbeat(p *Packet) {

}

func (client *Client) handlePing(p *Packet) {
	packet := &Packet{
		Ver: p.Ver,
		Mt:  MESSAGE_TYPE_PONG,
		Mid: 0,
		Sid: 0,
		Rid: 0,
	}

	client.sendChan <- packet
}

func (client *Client) handleRegister(p *Packet) {
	//从payload获取注册设备信息
	deviceInfo := DeviceInfo{}
	err := json.Unmarshal(p.Pl, &deviceInfo)
	if err != nil || deviceInfo.Token == "" {
		//输出注册失败信息
		packet := &Packet{
			Ver: p.Ver,
			Mt:  MESSAGE_TYPE_REGISTER_STATUS,
			Mid: 0,
			Sid: 0,
			Rid: 0,
			Pl:  buildResponseInfo(-1, "params decode err"),
		}
		client.sendChan <- packet
		return
	}

	c := client.server.GetClientByDt(deviceInfo.Token)
	if c == nil {
		//注册新的客户端
		client.server.RegisterClientByDt(client, deviceInfo.Token)
	} else {
		//有老的客户端
		if client == c {
			//同一个客户端 do nothing
		} else {
			//不是同一个客户端，注销之前的客户端
			client.server.UnRegisterClient(c.uid, deviceInfo.Token)
			//注册新的客户端
			client.server.RegisterClientByDt(client, deviceInfo.Token)
		}
	}

	if client.IsAuth() {
		client.server.RegisterClientByUid(client, client.uid)
	}

	//返回成功回执
	packet := &Packet{
		Ver: p.Ver,
		Mt:  MESSAGE_TYPE_REGISTER_STATUS,
		Mid: 0,
		Sid: 0,
		Rid: 0,
		Pl:  buildResponseInfo(0, ""),
	}
	client.sendChan <- packet
}

func (client *Client) handleAuth(p *Packet) {
	//获取鉴权信息
	authInfo := AuthInfo{}
	err := json.Unmarshal(p.Pl, &authInfo)
	if err != nil || authInfo.Uid == 0 || authInfo.Token == "" {
		//输出鉴权失败信息
		packet := &Packet{
			Ver: p.Ver,
			Mt:  MESSAGE_TYPE_AUTH_STATUS,
			Mid: 0,
			Sid: 0,
			Rid: 0,
			Pl:  buildResponseInfo(-1, "params decode err"),
		}
		client.sendChan <- packet
		return
	}

	//获取鉴权信息
	//判断鉴权通过
	if authInfo.Uid != 1 || authInfo.Token != "123" {
		packet := &Packet{
			Ver: p.Ver,
			Mt:  MESSAGE_TYPE_AUTH_STATUS,
			Mid: 0,
			Sid: 0,
			Rid: 0,
			Pl:  buildResponseInfo(-2, "auth failed"),
		}
		client.sendChan <- packet
	}

	//通过
	c := client.server.GetClientByUid(authInfo.Uid)
	if c == nil {
		//注册新的客户端
		client.server.RegisterClientByUid(client, authInfo.Uid)
	} else {
		//有老的客户端
		if client == c {
			//同一个客户端 do nothing
		} else {
			//不是同一个客户端，注销之前的客户端
			client.server.UnRegisterClient(authInfo.Uid, c.deviceToken)
			//注册新的客户端
			client.server.RegisterClientByDt(client, client.deviceToken)
			client.server.RegisterClientByUid(client, authInfo.Uid)
		}
	}

	atomic.StoreInt32(&client.authFlag, 1)

	//成功回执
	packet := &Packet{
		Ver: p.Ver,
		Mt:  MESSAGE_TYPE_AUTH_STATUS,
		Mid: 0,
		Sid: 0,
		Rid: 0,
		Pl:  buildResponseInfo(0, ""),
	}
	client.sendChan <- packet
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
