package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/256dpi/gomqtt/packet"

	"github.com/zxxf18/mqtt_client/protocol/mqtt"
	"github.com/zxxf18/mqtt_client/utils"
)

var (
	Interval = time.Millisecond * 100
	Topic    = "test2"
	QoS      = 0
)

func main() {
	err := sender()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func sender() error {
	cfg := mqtt.ClientInfo{
		Address:  "tcp://0.0.0.0:1883",
		Username: "test",
		Password: "test",
		ClientID: utils.UUID(),
	}

	cli := mqtt.NewDispatcher(cfg)
	cli.Start(nil)
	ticker := time.NewTicker(Interval)
	defer ticker.Stop()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)

	for {
		select {
		case <-ticker.C:
			pkt := packet.NewPublish()
			pkt.Message.Topic = Topic
			pkt.Message.QOS = packet.QOS(QoS)
			pkt.Message.Payload = []byte(time.Now().String())
			err := cli.Send(pkt)
			if err != nil {
				return fmt.Errorf("failed to publish: %s", err.Error())
			}
			fmt.Println("send success", string(pkt.Message.Payload))
		case <-sig:
			fmt.Println("os signal exit")
			return nil
		}
	}
}
