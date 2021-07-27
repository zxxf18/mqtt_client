package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/256dpi/gomqtt/packet"

	_ "github.com/go-sql-driver/mysql"

	"github.com/zxxf18/mqtt_client/config"
	"github.com/zxxf18/mqtt_client/protocol/mqtt"
	"github.com/zxxf18/mqtt_client/utils"
)

var (
	f string
)

func init() {
	flag.StringVar(&f, "f", "etc/config.yaml", "the configuration file")
}

func main() {
	err := sub()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}

func sub() error {
	flag.Parse()

	var cfg config.Config
	err := utils.LoadYAML(f, &cfg)
	if err != nil {
		return err
	}
	cfg.MQTT.ClientID = utils.UUID()

	db, err := NewDB(f)
	if err != nil {
		return err
	}
	defer db.Close()

	svr, err := NewServer(f, db)
	if err != nil {
		return err
	}
	defer svr.Close()
	go svr.Run()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)

	<-sig

	cli, err := mqtt.NewClient(cfg.MQTT.ClientInfo, &handler{db: db})
	if err != nil {
		return err
	}

	return cli.Close()
}

type handler struct {
	db *DB
}

func (hd *handler) ProcessPublish(publish *packet.Publish) error {
	//fmt.Println("ProcessPublish --> ", publish.Message.String())
	fmt.Println()
	fmt.Println(string(publish.Message.Payload))
	err := hd.db.Save(publish.Message.Topic, publish.Message.Payload)
	if err != nil {
		fmt.Println("failed to save content", err.Error())
	}
	return nil
}

func (hd *handler) ProcessPuback(puback *packet.Puback) error {
	fmt.Println("ProcessPuback --> ", puback.String())
	return nil
}

func (hd *handler) ProcessError(err error) {
	fmt.Println("ProcessError --> ", err.Error())
}
