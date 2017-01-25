package tcpserver

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
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

type Protocol interface {
	ReadPacket(conn net.Conn) (*Packet, error)
	WritePacket(conn net.Conn, p *Packet) error
	Serialize(p *Packet) []byte
	Unserialize(data []byte) (*Packet, error)
}

//消息结构体
type Packet struct {
	ver int32  //协议版本号
	mt  int32  //消息类型
	mid int64  //消息id
	sid int64  //发送者id
	rid int64  //接收者id
	ext []byte //附加属性字典
	pl  []byte //内容payload
}

type CustomProto struct {
}

func (proto *CustomProto) ReadPacket(conn net.Conn) (*Packet, error) {
	//先读取单条数据长度 2^32
	buf := make([]byte, 4)
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		fmt.Printf("read packet length error: %s\n", err.Error())
		return nil, err
	}

	var length int32
	buffer := bytes.NewBuffer(buf)
	binary.Read(buffer, binary.BigEndian, &length)

	//往后读取计算出长度字节
	buf = make([]byte, length)
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		fmt.Printf("read packet body error: %s\n", err.Error())
		return nil, err
	}

	return proto.Unserialize(buf)
}

func (proto *CustomProto) WritePacket(conn net.Conn, p *Packet) error {
	buf := proto.Serialize(p)
	n, err := conn.Write(buf)
	if err != nil {
		fmt.Printf("socket write error: %s\n", err.Error())
		return err
	}

	if n != len(buf) {
		fmt.Printf("socket write less: %d, %d\n", n, len(buf))
		return errors.New("socket write less")
	}

	return nil
}

func (proto *CustomProto) Serialize(p *Packet) []byte {
	//写入消息总长度 4 + 4 + 8 + 8 + 8 + 4 + 4 + len(ext) + len(payload)
	buffer := new(bytes.Buffer)

	var length, extLength, plLength int32

	extLength = int32(len(p.ext))
	plLength = int32(len(p.pl))
	length = 40 + extLength + plLength
	//写入长度
	binary.Write(buffer, binary.BigEndian, length)
	//写入协议版本号
	binary.Write(buffer, binary.BigEndian, p.ver)
	//写入消息类型
	binary.Write(buffer, binary.BigEndian, p.mt)
	//写入消息id
	binary.Write(buffer, binary.BigEndian, p.mid)
	//写入发送者id
	binary.Write(buffer, binary.BigEndian, p.sid)
	//写入接收者id
	binary.Write(buffer, binary.BigEndian, p.rid)
	//写入ext属性长度
	binary.Write(buffer, binary.BigEndian, extLength)
	//写入payload长度
	binary.Write(buffer, binary.BigEndian, plLength)
	buffer.Write(p.ext)
	buffer.Write(p.pl)

	buf := buffer.Bytes()

	return buf
}

func (proto *CustomProto) Unserialize(data []byte) (*Packet, error) {
	//先读取单条数据长度 2^32
	l := len(data)
	if l < 50 {
		return nil, errors.New("packet unserialize error")
	}

	p := Packet{}
	buffer := bytes.NewBuffer(data)
	//读取4字节协议版本号
	binary.Read(buffer, binary.BigEndian, &p.ver)
	//读取4字节消息类型
	binary.Read(buffer, binary.BigEndian, &p.mt)
	//读取8字节消息id
	binary.Read(buffer, binary.BigEndian, &p.mid)
	//读取8字节发送者id
	binary.Read(buffer, binary.BigEndian, &p.sid)
	//读取8字节接收者id
	binary.Read(buffer, binary.BigEndian, &p.rid)
	var extLength, plLength int32
	//读取4字节ext属性长度
	binary.Read(buffer, binary.BigEndian, &extLength)
	//读取4字节payload长度
	binary.Read(buffer, binary.BigEndian, &plLength)

	p.ext = make([]byte, extLength)
	p.pl = make([]byte, plLength)
	buffer.Read(p.ext)
	buffer.Read(p.pl)

	return &p, nil
}
