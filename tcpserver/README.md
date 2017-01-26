##tcpserver im框架##

![image](https://github.com/yybirdcf/go/raw/master/tcpserver/tcpserver.jpg)

####1.协议####

采用二进制自定义协议：

4字节表示packet长度 + packet<br />

packet = (<br />
 4字节包版本 + <br />
 4字节消息类型 + <br />
 8字节消息id + <br />
 8字节发送者id + <br />
 8字节接收者id + <br />
 4字节ext扩展信息长度 + <br />
 4字节payload信息长度 + <br />
 ext扩展数据 + <br />
 payload扩展信息<br />
)<br />