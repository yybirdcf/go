package tcpserver

import (
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync"
)

type TCPServer struct {
	address    string
	quit       chan bool
	sub        *Subscribe //订阅消息
	protocol   Protocol   //消息解析协议
	mutex      sync.Mutex
	uidClients map[int64]*Client  //用户id对应客户端映射表
	dtClients  map[string]*Client //设备token对应客户端映射表
	inChan     chan *Packet       //客户端写入到服务器
	outChan    chan *Packet       //服务器下发到客户端
}

func NewTCPServer(listenaddr string, nsqaddr string) *TCPServer {
	server := &TCPServer{
		address:    listenaddr,
		quit:       make(chan bool),
		protocol:   &CustomProto{},
		uidClients: make(map[int64]*Client),
		dtClients:  make(map[string]*Client),
		inChan:     make(chan *Packet, 1024),
		outChan:    make(chan *Packet, 1024),
	}

	server.sub = NewSubscribe(server.protocol, nsqaddr, MESSAGE_TOPIC_DISPATCH, MESSAGE_CHANNEL_DISPATCH_IM, server.outChan)

	go server.inLoop()
	go server.outLoop()

	return server
}

func (server *TCPServer) GetClientByUid(uid int64) *Client {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	if client, ok := server.uidClients[uid]; ok {
		return client
	}

	return nil
}

func (server *TCPServer) GetClientByDt(dt string) *Client {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	if client, ok := server.dtClients[dt]; ok {
		return client
	}

	return nil
}

func (server *TCPServer) RegisterClientByUid(client *Client, uid int64) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	if _, ok := server.uidClients[uid]; !ok {
		server.uidClients[uid] = client
	}
}

func (server *TCPServer) RegisterClientByDt(client *Client, dt string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	if _, ok := server.dtClients[dt]; !ok {
		server.dtClients[dt] = client
	}
}

func (server *TCPServer) UnRegisterClient(uid int64, dt string) {
	server.mutex.Lock()
	defer server.mutex.Unlock()

	if _, ok := server.uidClients[uid]; ok {
		delete(server.uidClients, uid)
	}

	if _, ok := server.dtClients[dt]; ok {
		delete(server.dtClients, dt)
	}
}

func (server *TCPServer) Close() {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for k, v := range server.uidClients {
		v.Close()
		delete(server.uidClients, k)
	}

	for k, _ := range server.dtClients {
		delete(server.dtClients, k)
	}

	server.sub.Close()
	server.quit <- true
	close(server.quit)
	close(server.inChan)
	close(server.outChan)
	server.sub.Close()
}

func (server *TCPServer) Serve() error {
	listener, err := net.Listen("tcp", server.address)
	if err != nil {
		fmt.Printf("error listen tcp: %s, error: %s\n", server.address, err.Error())
		return err
	}

	fmt.Printf("listen tcp: %s\n", server.address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				fmt.Printf("NOTICE: temporary Accept() failure - %s\n", err)
				runtime.Gosched() // 是临时的错误, 暂停一下继续
				continue
			}
			// theres no direct way to detect this error because it is not exposed
			if !strings.Contains(err.Error(), "use of closed network connection") {
				fmt.Printf("ERROR: listener.Accept() - %s\n", err)
			}
			return err
		}
		//启动一个线程, 交给 handler 处理, 这里使用的是 one connect per thread 模式
		//因为golang的特性, one connect per thread 模式 实际上是  one connect per goroutine
		go server.handle(conn)
	}
}

func (server *TCPServer) handle(conn net.Conn) {
	//创建客户端
	client := NewClient(server, conn)
	//运行客户端
	client.Do()
}

//处理客户端写入的消息
func (server *TCPServer) inLoop() {
	for {
		select {
		case <-server.quit:
			return
		case p := <-server.inChan:
			//写到nsq分发
			server.sub.Publish(MESSAGE_TOPIC_LOGIC, p)
		}
	}
}

//服务端将信息写到客户端
func (server *TCPServer) outLoop() {
	for {
		select {
		case <-server.quit:
			return
		case p := <-server.outChan:
			if c, ok := server.uidClients[p.Rid]; ok {
				c.sendChan <- p
			}
		}
	}
}
