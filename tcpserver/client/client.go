package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/tcpserver"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		listenAddr = flag.String("listen.addr", ":12000", "tcp listen address")
	)
	flag.Parse()

	conn, err := net.Dial("tcp", *listenAddr)
	if err != nil {
		fmt.Println(err)
	}

	proto := &tcpserver.CustomProto{}

	//绑定设备信息
	deviceToken := tcpserver.DeviceInfo{
		Token: "12345678",
	}

	bs, err := json.Marshal(deviceToken)
	if err != nil {
		fmt.Println(err)
	}

	p := &tcpserver.Packet{
		Ver: 1,
		Mt:  tcpserver.MESSAGE_TYPE_REGISTER,
		Mid: 0,
		Sid: 0,
		Rid: 0,
		Pl:  bs,
	}

	proto.WritePacket(conn, p)

	// go func() {
	// 	for {
	// 		time.Sleep(time.Second * 3)
	//
	// 		p := &tcpserver.Packet{
	// 			Ver: 1,
	// 			Mt:  tcpserver.MESSAGE_TYPE_PING,
	// 			Mid: 0,
	// 			Sid: 0,
	// 			Rid: 0,
	// 		}
	//
	// 		proto.WritePacket(conn, p)
	// 	}
	// }()

	resp := tcpserver.ResponseInfo{}

	go func() {
		for {
			p, err := proto.ReadPacket(conn)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Printf("%+v\n", *p)
			//注册设备信息
			if p.Mt == tcpserver.MESSAGE_TYPE_REGISTER_STATUS {
				json.Unmarshal(p.Pl, &resp)
				fmt.Printf("%+v\n", resp)
				if resp.Status == 0 {
					//鉴权
					authInfo := tcpserver.AuthInfo{
						Uid:   1,
						Token: "123",
					}

					bs, err := json.Marshal(authInfo)
					if err != nil {
						fmt.Println(err)
					}

					p := &tcpserver.Packet{
						Ver: 1,
						Mt:  tcpserver.MESSAGE_TYPE_AUTH,
						Mid: 0,
						Sid: 0,
						Rid: 0,
						Pl:  bs,
					}

					proto.WritePacket(conn, p)
				}
			} else if p.Mt == tcpserver.MESSAGE_TYPE_AUTH_STATUS {
				json.Unmarshal(p.Pl, &resp)
				fmt.Printf("%+v\n", resp)
			}
		}
	}()

	errc := make(chan error)

	// Interrupt handler.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	fmt.Printf("exit: %v", <-errc)
}
