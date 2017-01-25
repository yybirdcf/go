package tcpserver

import (
	"fmt"
	"os"

	nsq "github.com/nsqio/go-nsq"
)

var (
	MESSAGE_TOPIC_LOGIC         = "message_topic_logic"         //后端消息处理服务
	MESSAGE_CHANNEL_LOGIC_IM    = "message_channel_logic_im"    //处理聊天消息
	MESSAGE_TOPIC_DISPATCH      = "message_topic_dispatch"      //服务器分发消息
	MESSAGE_CHANNEL_DISPATCH_IM = "message_channel_dispatch_im" //分发聊天消息
)

type Subscribe struct {
	protocol Protocol      //消息解析协议
	producer *nsq.Producer //发送消息到逻辑处理层
	consumer *nsq.Consumer //分发消息到客户端
	outChan  chan *Packet
}

func NewSubscribe(protocol Protocol, nsqaddr string, topic string, c string, out chan *Packet) *Subscribe {
	cfg := nsq.NewConfig()
	producer, err := nsq.NewProducer(nsqaddr, cfg)
	if err != nil {
		fmt.Printf("nsq producer error: %s\n", err.Error())
		os.Exit(1)
	}

	consumer, err := nsq.NewConsumer(topic, c, cfg)
	if err != nil {
		fmt.Printf("nsq consumer error: %s\n", err.Error())
		os.Exit(1)
	}

	sub := &Subscribe{
		protocol: protocol,
		producer: producer,
		consumer: consumer,
		outChan:  out,
	}

	go func() {
		// 设置消息处理函数
		consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
			p, err := sub.protocol.Unserialize(message.Body)
			if err == nil {
				sub.outChan <- p
			}
			return nil
		}))

		// 连接到单例nsqd
		if err := consumer.ConnectToNSQD(nsqaddr); err != nil {
			fmt.Printf("nsq consumer ConnectToNSQD error: %s\n", err.Error())
			os.Exit(1)
		}
		<-consumer.StopChan
	}()

	return sub
}

func (sub *Subscribe) Publish(topic string, p *Packet) {
	sub.producer.Publish(topic, sub.protocol.Serialize(p))
}

func (sub *Subscribe) Close() {
	sub.consumer.StopChan <- 1
}
