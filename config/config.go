package config

import (
	"github.com/zxxf18/mqtt_client/protocol/mqtt"
)

type Config struct {
	MQTT struct {
		mqtt.ClientInfo `yaml:",inline" json:",inline"`
	} `yaml:"mqtt" json:"mqtt"`
}
